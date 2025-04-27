package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

const webhookURL = "https://discord.com/api/webhooks/1365944421761159208/9xcOtblEKZpLi0CbH8YxU_99Y5fhHrxsJMu7QUTBmyQYqrNxq5kuzZ_iXOoVpJRrPVY6"

var (
	mainToken string
	// Ultra-optimized transport for minimal latency
	transport = &http.Transport{
		MaxIdleConns:          1000,
		MaxIdleConnsPerHost:   250,
		MaxConnsPerHost:       250,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   3 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableCompression:    true,
		DisableKeepAlives:     false,
		ForceAttemptHTTP2:     true,
	}
	client = &http.Client{
		Transport: transport,
		Timeout:   3 * time.Second, // Ultra-fast timeout
	}
	// Pre-compiled regex for better performance
	giftRegex       = regexp.MustCompile(`(?i)(?:https?:\/\/)?(?:discord(?:app)?\.(?:com|gift)\/(?:gift|gifts)\/|discord\.gift\/)([A-Za-z0-9-]+)`)
	webhookClient   = &http.Client{Transport: transport, Timeout: 3 * time.Second}
	webhookChan     = make(chan webhookData, 200) // Increased buffer
	claimRateLimiter = make(chan struct{}, 20)    // Increased concurrent claims
	// Frequently accessed strings
	giftString1     = "discord.gift/"
	giftString2     = "discord.com/gift/"
)

type webhookData struct {
	userID      string
	messageLink string
	status      string
	elapsedMs   float64
}

func init() {
	// Start multiple webhook workers for faster processing
	for i := 0; i < 3; i++ {
		go webhookWorker()
	}
}

func webhookWorker() {
	for data := range webhookChan {
		// Fast webhook processing with minimal formatting
		embed := fmt.Sprintf(`{"embeds":[{"title":"🎁 Gift Link Detected!","color":%d,"description":"**User:** <@%s>\n**Gift Link:** [Click Here](%s)\n**Status:** %s\n**Time Taken:** %.2fms"}]}`,
			getEmbedColor(data.status),
			data.userID,
			data.messageLink,
			data.status,
			data.elapsedMs)

		req, err := http.NewRequest("POST", webhookURL, bytes.NewReader([]byte(embed)))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		
		_, _ = webhookClient.Do(req) // Ignoring errors for speed
	}
}

func loadTokens(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tokens []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		t := strings.TrimSpace(scanner.Text())
		if t != "" {
			tokens = append(tokens, t)
		}
	}
	return tokens, scanner.Err()
}

// Fast gift code extraction
func extractGiftCode(content string) string {
	if content == "" {
		return ""
	}
	
	// Direct string search first (faster than regex)
	if idx := strings.Index(strings.ToLower(content), giftString1); idx != -1 {
		start := idx + len(giftString1)
		if start < len(content) {
			end := start
			for end < len(content) && !isInvalidCodeChar(content[end]) {
				end++
			}
			if end > start && end-start >= 5 && end-start <= 24 {
				return content[start:end]
			}
		}
	}
	
	if idx := strings.Index(strings.ToLower(content), giftString2); idx != -1 {
		start := idx + len(giftString2)
		if start < len(content) {
			end := start
			for end < len(content) && !isInvalidCodeChar(content[end]) {
				end++
			}
			if end > start && end-start >= 5 && end-start <= 24 {
				return content[start:end]
			}
		}
	}
	
	// Fallback to regex only if needed
	match := giftRegex.FindStringSubmatch(content)
	if len(match) > 1 {
		return match[1]
	}
	
	return ""
}

// Helper to determine if a character is valid in a gift code
func isInvalidCodeChar(c byte) bool {
	return !(('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9') || c == '-')
}

func claimGift(code string) (bool, error) {
	// Use rate limiter to prevent Discord from blocking requests
	select {
	case claimRateLimiter <- struct{}{}:
		defer func() { <-claimRateLimiter }()
	default:
		// If channel is full, still proceed but don't block
		go func() { 
			claimRateLimiter <- struct{}{} 
			time.Sleep(100 * time.Millisecond)
			<-claimRateLimiter
		}()
	}
	
	url := fmt.Sprintf("https://discord.com/api/v9/entitlements/gift-codes/%s/redeem", code)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", mainToken)
	req.Header.Set("Content-Type", "application/json")
	
	// Adding optimized headers for faster processing
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Origin", "https://discord.com")
	req.Header.Set("Referer", "https://discord.com/channels/@me")
	
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200, nil
}

func getEmbedColor(status string) int {
	if status == "VALID" {
		return 0x57F287
	}
	return 0xED4245
}

func queueWebhook(userID, messageLink, status string, elapsedMs float64) {
	select {
	case webhookChan <- webhookData{userID, messageLink, status, elapsedMs}:
		// Successfully queued
	default:
		// Non-blocking if channel is full
	}
}

// Ultra-fast message processing
func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Skip own messages
	if m.Author.ID == s.State.User.ID {
		return
	}
	
	start := time.Now() // Start timing immediately
	
	// Ultra-fast string search instead of regex when possible
	content := strings.ToLower(m.Content)
	hasGift := strings.Contains(content, "discord.gift") || strings.Contains(content, "discord.com/gift")
	
	var code string
	if hasGift {
		code = extractGiftCode(m.Content)
	}
	
	// Only check embeds if we didn't find a code and there are embeds
	if code == "" && len(m.Embeds) > 0 {
		for _, embed := range m.Embeds {
			// Check likely places for gift links
			if code = extractGiftCode(embed.Description); code != "" {
				break
			}
			if code = extractGiftCode(embed.Title); code != "" {
				break
			}
			if code = extractGiftCode(embed.URL); code != "" {
				break
			}
			
			// Only check fields if we must
			for _, field := range embed.Fields {
				if code = extractGiftCode(field.Value); code != "" {
					break
				}
			}
			if code != "" {
				break
			}
		}
	}
	
	if code != "" {
		// Process gift claiming in a goroutine
		go func(startTime time.Time, giftCode, userID, guildID, channelID, messageID string) {
			success, err := claimGift(giftCode)
			elapsed := time.Since(startTime).Seconds() * 1000
			
			status := "INVALID"
			if success {
				status = "VALID"
			}
			
			fmt.Printf("🎯 code %s | %s | %.1fms\n", giftCode, status, elapsed)
			
			link := fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildID, channelID, messageID)
			queueWebhook(userID, link, status, elapsed)
			
			if err != nil {
				fmt.Printf("❌ Error: %s\n", err)
			}
		}(start, code, m.Author.ID, m.GuildID, m.ChannelID, m.ID)
	}
}

func loginAlt(token string, wg *sync.WaitGroup) {
	defer wg.Done()
	
	// Create optimized session
	dg, err := discordgo.New(token)
	if err != nil {
		fmt.Println("⚠️ Error creating session:", err)
		return
	}
	
	// Minimal intents for speed
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages | discordgo.IntentsMessageContent
	
	dg.Identify.Properties = discordgo.IdentifyProperties{
		OS:      "Windows",
		Browser: "Discord Client",
		Device:  "Desktop",
	}
	
	dg.AddHandler(onMessageCreate)
	
	// Optimize connection
	dg.Client.Transport = transport
	dg.Client.Timeout = 10 * time.Second
	
	if err := dg.Open(); err != nil {
		fmt.Println("⚠️ Failed to open connection:", err)
		return
	}
	
	fmt.Printf("✅ Logged in as %s\n", dg.State.User.ID)
}

func main() {
	// Initialize rate limiter channel
	for i := 0; i < 20; i++ {
		claimRateLimiter <- struct{}{}
	}
	
	mains, err := loadTokens("main.txt")
	if err != nil || len(mains) == 0 {
		fmt.Println("❌ could not load main token")
		return
	}
	mainToken = mains[0]
	
	alts, err := loadTokens("alt.txt")
	if err != nil {
		fmt.Println("❌ could not load alt tokens")
		return
	}
	
	fmt.Println("🚀 Discord Gift Sniper Launching...")
	fmt.Println("💨 Optimized for ultra-fast gift claiming")
	
	var wg sync.WaitGroup
	for _, t := range alts {
		wg.Add(1)
		go loginAlt(t, &wg)
	}
	wg.Wait()
	fmt.Println("⚡ All systems ready. Hunting for gifts...")
	
	// Keep program running
	select {}
}
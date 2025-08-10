![Banner](https://iili.io/FLPMSwJ.md.gif)

<h1 align="center">🎯 Ultra-Fast Discord Gift Sniper</h1>
<p align="center">
  <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=for-the-badge&logo=go" alt="Go Badge">
  <img src="https://img.shields.io/badge/Speed-Ultra%20Fast-ff0000?style=for-the-badge&logo=rocket" alt="Speed Badge">
  <img src="https://img.shields.io/badge/Status-Active-success?style=for-the-badge&logo=github" alt="Status Badge">
  <img src="https://img.shields.io/github/license/bruciaa/Best-sniper-ever?style=for-the-badge" alt="License Badge">
</p>

<p align="center">
  <b>A next-gen Discord Nitro Gift Code Sniper built with Go for ultra-low latency and maximum claiming efficiency.</b>
</p>

---

## 🚀 Features

- ⚡ **Ultra-Low Latency** — Highly optimized HTTP transport for instant responses.
- 🔍 **Smart Gift Code Detection** — Direct string search with regex fallback for accuracy.
- 📢 **Webhook Notifications** — Instant updates with claim status and timing.
- 🛠 **Multi-Account Support** — Load multiple alt tokens for faster claiming.
- 📊 **Optimized Concurrency** — Rate-limited parallel claims to avoid bans.
- 📦 **Lightweight & Efficient** — Pure Go, no heavy dependencies.

---

## 📂 Project Structure

```
.
├── main.go       # Main sniper logic
├── main.txt      # File containing your main Discord token
├── alt.txt       # File containing your alt account tokens
```

---

## ⚙️ Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/bruciaa/Best-sniper-ever.git
   cd Best-sniper-ever
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Add your tokens**
   - `main.txt` → Your main Discord account token (first line only)
   - `alt.txt` → List of alt account tokens (one per line)

4. **Run the sniper**
   ```bash
   go run main.go
   ```

---

## 🔧 Configuration

- **Webhook URL** — Update `webhookURL` in `main.go` to your Discord webhook.
- **Concurrency** — Adjust `claimRateLimiter` for more/fewer parallel claims.
- **Timeouts** — Tweak `http.Transport` and `http.Client` for speed.

---

## ⚠️ Disclaimer

> **Warning:** This tool interacts with Discord's API in a way that may violate Discord's Terms of Service.  
> Use at your own risk — accounts and IPs may be banned.

This project is for **educational purposes only**.

---

## 📜 License

MIT License © 2025 BRUCE

# Docker Deployment

Run Picobot as a Docker container — one command to start.

## Quick Start

### Option 1: Docker Compose (Recommended)

```sh
# 1. Create .env with your API key and settings
nano docker/.env

# 2. Start
docker compose -f docker/docker-compose.yml up -d

# 3. Check logs
docker compose -f docker/docker-compose.yml logs -f
```

### Option 2: Docker Run

```sh
# Build the image
docker build -f docker/Dockerfile -t picobot .

# Run with environment variables
docker run -d \
  --name picobot \
  --restart unless-stopped \
  -e OPENROUTER_API_KEY="sk-or-v1-YOUR_KEY" \
  -e PICOBOT_MODEL="google/gemini-2.5-flash" \
  -e TELEGRAM_BOT_TOKEN="123456:ABC..." \
  -e TELEGRAM_ALLOW_FROM="8881234567" \
  -v ./picobot-data:/home/picobot/.picobot \
  picobot
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OPENROUTER_API_KEY` | Yes | — | Your OpenRouter API key |
| `PICOBOT_MODEL` | No | `google/gemini-2.5-flash` | LLM model to use |
| `TELEGRAM_BOT_TOKEN` | No | — | Telegram bot token from @BotFather |
| `TELEGRAM_ALLOW_FROM` | No | — | Comma-separated Telegram user IDs |

## Data Persistence

All data is stored in the `picobot-data` Docker volume:
- `config.json` — configuration
- `workspace/` — bootstrap files, memory, skills

Data persists across container restarts and rebuilds.

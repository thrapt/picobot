# How to Start Using Picobot

## Prerequisites

- **Go 1.21+** installed ([download](https://go.dev/dl/))
- An **OpenRouter API key** ([get one here](https://openrouter.ai/keys))

## Step 1: Build

```sh
git clone <repo-url>
cd picobot-main
go build -o picobot ./cmd/picobot
```

## Step 2: Onboard

Run the onboard command to create the config and workspace:

```sh
./picobot onboard
```

This creates:
- `~/.picobot/config.json` — your configuration file
- `~/.picobot/workspace/` — the agent's workspace with bootstrap files:
  - `SOUL.md` — agent personality and values
  - `AGENTS.md` — agent instructions and guidelines
  - `USER.md` — your profile (customize this!)
  - `TOOLS.md` — documentation of all available tools
  - `HEARTBEAT.md` — periodic tasks
  - `memory/MEMORY.md` — long-term memory
  - `skills/example/SKILL.md` — example skill

## Step 3: Configure API Key

Edit `~/.picobot/config.json` and replace the placeholder API key:

```sh
# Open in your editor
nano ~/.picobot/config.json
```

Change `"sk-or-v1-REPLACE_ME"` to your actual OpenRouter API key.

Also set your preferred model (e.g., `google/gemini-2.5-flash`):

```json
{
  "agents": {
    "defaults": {
      "model": "google/gemini-2.5-flash"
    }
  },
  "providers": {
    "openrouter": {
      "apiKey": "sk-or-v1-YOUR_ACTUAL_KEY",
      "apiBase": "https://openrouter.ai/api/v1"
    }
  }
}
```

## Step 4: Customize Your Profile

Edit `~/.picobot/workspace/USER.md` to fill in your name, timezone, preferences, etc. This helps the agent personalize its responses.

## Step 5: Try It!

### Single-shot query

```sh
./picobot agent -m "Hello, what tools do you have?"
```

### Use a specific model

```sh
./picobot agent -M "google/gemini-2.5-flash" -m "What is 2+2?"
```

### Start the gateway (long-running mode)

```sh
./picobot gateway
```

This starts the agent loop, heartbeat, and any enabled channels (e.g., Telegram).

## CLI Commands

| Command | Description |
|---------|-------------|
| `picobot version` | Print version |
| `picobot onboard` | Create default config and workspace |
| `picobot agent -m "..."` | Run a single-shot agent query |
| `picobot agent -M model -m "..."` | Query with a specific model |
| `picobot gateway` | Start long-running gateway |
| `picobot memory read today` | Read today's memory notes |
| `picobot memory read long` | Read long-term memory |
| `picobot memory append today -c "..."` | Append to today's notes |
| `picobot memory append long -c "..."` | Append to long-term memory |
| `picobot memory write long -c "..."` | Overwrite long-term memory |
| `picobot memory recent -days 7` | Show recent 7 days' notes |
| `picobot memory rank -q "query"` | Rank memories by relevance |

## Available Tools

The agent has access to 11 tools:

| Tool | Purpose |
|------|---------|
| `message` | Send messages to channels |
| `filesystem` | Read, write, list files |
| `exec` | Run shell commands |
| `web` | Fetch web content from URLs |
| `spawn` | Spawn background subagent |
| `cron` | Schedule cron jobs |
| `write_memory` | Persist information to memory |
| `create_skill` | Create a new skill |
| `list_skills` | List available skills |
| `read_skill` | Read a skill's content |
| `delete_skill` | Delete a skill |

## Setting Up Telegram (BotFather Guide)

To chat with Picobot on Telegram, you need to create a bot via **@BotFather**.

### 1. Open BotFather

Open Telegram and search for [@BotFather](https://t.me/BotFather), or click the link directly. This is Telegram's official bot for creating and managing bots.

### 2. Create a New Bot

Send the command:

```
/newbot
```

BotFather will ask you two questions:

1. **Bot name** — A display name (e.g., `My Picobot`)
2. **Bot username** — A unique username ending in `bot` (e.g., `my_picobot_bot`)

### 3. Copy the Token

After creation, BotFather will reply with a message like:

```
Done! Congratulations on your new bot. You will find it at t.me/my_picobot_bot.
Use this token to access the HTTP API:
123456789:ABCdefGHIjklMNOpqrsTUVwxyz
```

Copy the token — you'll need it in the next step.

### 4. Get Your Telegram User ID

To restrict who can talk to your bot, you need your numeric Telegram user ID.

Send a message to [@userinfobot](https://t.me/userinfobot) on Telegram — it will reply with your user ID (a number like `8881234567`).

### 5. Configure Picobot

Edit `~/.picobot/config.json` and add your Telegram settings:

```json
{
  "channels": {
    "telegram": {
      "enabled": true,
      "token": "123456789:ABCdefGHIjklMNOpqrsTUVwxyz",
      "allowFrom": ["8881234567"]
    }
  }
}
```

| Field | Description |
|-------|-------------|
| `enabled` | Set to `true` to activate the Telegram channel |
| `token` | The bot token from BotFather |
| `allowFrom` | List of user IDs allowed to chat. Empty `[]` = anyone can use it |

### 6. Start the Gateway

```sh
./picobot gateway
```

Now open Telegram, find your bot by its username, and send it a message. Picobot will respond!

### Optional: Customize Your Bot in BotFather

You can also send these commands to @BotFather to polish your bot:

| Command | What it does |
|---------|-------------|
| `/setdescription` | Short description shown on the bot's profile |
| `/setabouttext` | "About" text in the bot info page |
| `/setuserpic` | Upload a profile photo for your bot |
| `/setcommands` | Set the bot's command menu (e.g., `/start`) |
| `/mybots` | Manage all your bots |

---

## Next Steps

- Edit `SOUL.md` to change the agent's personality
- Edit `AGENTS.md` to add custom instructions
- Ask the agent to create skills for tasks you do often
- Enable Telegram in `config.json` to chat with your bot on mobile
- See [CONFIG.md](CONFIG.md) for all configuration options

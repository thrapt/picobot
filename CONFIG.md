# Configuration Reference

Picobot is configured via `~/.picobot/config.json`. Run `picobot onboard` to generate the default config.

## Full Default Config

```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.picobot/workspace",
      "model": "stub-model",
      "maxTokens": 8192,
      "temperature": 0.7,
      "maxToolIterations": 100,
      "heartbeatIntervalS": 60
    }
  },
  "channels": {
    "telegram": {
      "enabled": false,
      "token": "",
      "allowFrom": []
    }
  },
  "providers": {
    "openrouter": {
      "apiKey": "sk-or-v1-REPLACE_ME",
      "apiBase": "https://openrouter.ai/api/v1"
    }
  }
}
```

---

## agents.defaults

Agent behavior settings.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `workspace` | string | `~/.picobot/workspace` | Path to the agent's workspace directory. Contains bootstrap files, memory, and skills. |
| `model` | string | `stub-model` | Default LLM model to use. Set to a real model like `google/gemini-2.5-flash`. Can be overridden with the `-M` flag. |
| `maxTokens` | int | `8192` | Maximum tokens for LLM responses. |
| `temperature` | float | `0.7` | LLM temperature (0.0 = deterministic, 1.0 = creative). |
| `maxToolIterations` | int | `100` | Maximum number of tool-calling iterations per request. Prevents infinite loops. |
| `heartbeatIntervalS` | int | `60` | How often (in seconds) the heartbeat checks `HEARTBEAT.md` for periodic tasks. Only used in gateway mode. |

### Model Priority

The model is resolved in this order:
1. **CLI flag** (`-M` / `--model`)
2. **Config** (`agents.defaults.model`)
3. **Provider default** (fallback)

### Example

```json
{
  "agents": {
    "defaults": {
      "workspace": "/home/user/.picobot/workspace",
      "model": "google/gemini-2.5-flash",
      "maxTokens": 16384,
      "temperature": 0.5,
      "maxToolIterations": 200,
      "heartbeatIntervalS": 120
    }
  }
}
```

---

## providers

LLM provider configuration. Picobot supports two providers. The first one with a valid config is used.

### providers.openrouter

Connect to [OpenRouter](https://openrouter.ai) (cloud).

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `apiKey` | string | *(required)* | Your OpenRouter API key. Get one at https://openrouter.ai/keys |
| `apiBase` | string | `https://openrouter.ai/api/v1` | API base URL. Change if using a proxy or compatible endpoint. |

```json
{
  "providers": {
    "openrouter": {
      "apiKey": "sk-or-v1-your-key-here",
      "apiBase": "https://openrouter.ai/api/v1"
    }
  }
}
```

### providers.ollama

Connect to a local [Ollama](https://ollama.ai) instance (self-hosted).

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `apiKey` | string | *(not needed)* | Not required for local Ollama. |
| `apiBase` | string | `http://localhost:11434/v1` | URL of your Ollama instance. |

```json
{
  "providers": {
    "ollama": {
      "apiKey": "",
      "apiBase": "http://localhost:11434/v1"
    }
  }
}
```

### Provider Selection

Picobot selects the provider in this order:
1. **OpenRouter** — if `openrouter.apiKey` is set and non-empty
2. **Ollama** — if `ollama.apiBase` is set and non-empty
3. **Stub** — fallback (echoes back your message, for testing)

To switch providers, remove or clear the one you don't want:

```json
{
  "providers": {
    "ollama": {
      "apiBase": "http://localhost:11434/v1"
    }
  }
}
```

---

## channels

Chat channel integrations. Currently supports Telegram.

### channels.telegram

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `false` | Set to `true` to start the Telegram bot. |
| `token` | string | `""` | Your Telegram Bot token from [@BotFather](https://t.me/BotFather). |
| `allowFrom` | string[] | `[]` | List of allowed Telegram user IDs. Empty = allow all. |

```json
{
  "channels": {
    "telegram": {
      "enabled": true,
      "token": "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
      "allowFrom": ["8881234567"]
    }
  }
}
```

---

## Workspace Files

The workspace directory (default `~/.picobot/workspace`) contains files that shape agent behavior:

| File | Purpose | Who edits |
|------|---------|-----------|
| `SOUL.md` | Agent personality, values, communication style | You (once) |
| `AGENTS.md` | Agent instructions, rules, guidelines | You (once) |
| `USER.md` | Your profile — name, timezone, preferences | You (once) |
| `TOOLS.md` | Tool reference documentation | You (once) |
| `HEARTBEAT.md` | Periodic tasks checked every `heartbeatIntervalS` seconds | You / Agent |
| `memory/MEMORY.md` | Long-term memory | Agent (via write_memory tool) |
| `memory/YYYY-MM-DD.md` | Daily notes | Agent (via write_memory tool) |
| `skills/` | Skill packages | Agent (via skill tools) or you manually |

---

## Example: Minimal Production Config

```json
{
  "agents": {
    "defaults": {
      "workspace": "/home/user/.picobot/workspace",
      "model": "google/gemini-2.5-flash",
      "maxTokens": 8192,
      "temperature": 0.7,
      "maxToolIterations": 200,
      "heartbeatIntervalS": 60
    }
  },
  "channels": {
    "telegram": {
      "enabled": true,
      "token": "YOUR_TELEGRAM_BOT_TOKEN",
      "allowFrom": ["YOUR_TELEGRAM_USER_ID"]
    }
  },
  "providers": {
    "openrouter": {
      "apiKey": "sk-or-v1-YOUR_KEY",
      "apiBase": "https://openrouter.ai/api/v1"
    }
  }
}
```

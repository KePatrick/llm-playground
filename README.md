# LLMplayground (go-clean-architecture)
中文版本README請查看: [README_zh.md](https://github.com/KePatrick/llm-playground/blob/main/README_zh.md)
## Project Overview

### Introduction

This is a practice project developed during spare time to learn and implement the following goals:

- Familiarize with **Go programming language**
- Master the practice of **Clean Architecture**
- Understand and implement **OpenAI-compatible API function call process**
- Build a chatbot assistant with memory and tool-calling support

The project focuses on structural clarity and functional correctness, and is intended for learning and reference purposes only. Feedback and contributions are welcome.

---

## Implemented Features

- Streaming response
- Conversation memory storage (Redis optional, defaults to local)
- Input/output logging (MySQL optional, defaults to local)
- Tool calling (Function Calling)
- Custom model and module configuration via JSON files

Note: Currently, when using the OpenAI API with streaming responses and opting to display token usage, the response behaves abnormally for unknown reasons. As a result, if the OpenAI API is used, token usage cannot be recorded.

---

## Clean Architecture Layering

This project follows the Clean Architecture structure for clarity and maintainability. The layered concept is:

| Layer | Folder | Responsibility |
|-------|--------|----------------|
| **Domain** | `internal/domain/entity`, `internal/domain/repository` | Defines business entities and contracts, independent of frameworks |
| **Use Case** | `internal/usecase` | Implements application logic (chat flow, tool selection), interacts with domain objects |
| **Interface Adapters** | `internal/gateway/http` | Controller logic, handles HTTP request and response conversion |
| **Frameworks & Drivers** | `internal/infra`, `scripts/` | Database, Redis, API implementations |
| **Entry Point & Configs** | `cmd/app`, `configs/` | Application initialization, config loading and service start-up |
| **Utility Packages** | `pkg/` | Tools and utilities outside core logic |

**Dependency Rule**
- **Principle**: Dependency direction must point inward only. Inner layers should not depend on outer layers.
  - Example: HTTP controllers (outer) may call Use Cases (inner), but Use Cases must not import controllers.
- **Implementation**:
  - Interfaces define contracts. Inner layers depend only on interfaces.
  - Outer layers (e.g., `main.go` or DI containers) inject concrete implementations into inner layers.
- **Benefits**:
  - **Decoupling**: Inner layers are unaffected by framework or tool changes.
  - **Replaceability**: Technology stack can be swapped without changing core logic.
  - **Testability**: Inner logic can be tested using mocks or fakes.

---

## System Requirements

- OpenAI-compatible API key (model must support function call)
- Custom tool scripts
- MySQL (optional)
- Redis (optional)

---

## Pre-Setup

### 1. Clone the repository

```bash
git clone https://github.com/yourusername/your-repo.git
cd your-repo
```

### 2. Configure config files

#### `configs/options.json` (Required)

```json
{
 "selectApi": "deepseek-chat",
 "sysPrompt": "you are a assistant about this project: llmplayground",
 "relationDatabase": false, 
 "redis": false 
}
```
Notes:

- selectApi: Corresponds to the API selection in configs/api.json
- sysPrompt: System prompt
- relationDatabase: Whether to enable database (only supports MySQL)
- redis: Whether to enable Redis
> Set `relationDatabase` and `redis` to `true` if needed, and configure `configs/database.json`, `configs/redis.json`.

#### `configs/api.json` (Required)

```json
{
  "deepseek-chat": {
    "apiKey": "your-api-key",
    "model": "deepseek-chat",
    "apiUrl": "https://api.deepseek.com/chat/completions"
  }
}
```

#### `configs/tools.json` (Optional)

```json
[
  {
    "type": "function",
    "function": {
      "name": "fetchProjectInfo",
      "description": "Query project info",
      "parameters": {
        "type": "object",
        "properties": {
          "languege": {
            "type": "string",
            "description": "Language: zh, en"
          }
        },
        "required": ["languege"]
      }
    },
    "script": "fetchProjectInfo"
  }
]
```

#### `configs/database.json` (Required if using MySQL)

```json
{
  "driver": "mysql",
  "name": "myDb",
  "url": "localhost:3306",
  "user": "user",
  "password": "password"
}
```
Notes:

- driver: Currently only supports MySQL

#### `configs/redis.json` (Required if using Redis)

```json
{
  "addr": "localhost:6379",
  "password": "",
  "db": 0
}
```

---

## Build and Run

```bash
go build -o app ./cmd/app
./app
```

- Default HTTP server listens on `localhost:8080`

---

## Usage

- Open browser and go to `localhost:8080` to chat
- Chat session ID is stored in browser session (cleared on close)
- If you want to clear memory, just open another window 
---

## Project Structure Summary

```
├── cmd/app              # Main app entry point
├── configs/             # JSON config files
├── internal/
│   ├── config/          # Config parser
│   ├── domain/          # Entities and interfaces
│   │   ├── entity/
│   │   ├── repository/
│   │   └── service/
│   ├── usecase/         # Application logic
│   ├── gateway/http/    # HTTP controllers and DTOs
│   └── infra/           # External dependencies: DB, Redis, LLM
├── local/               # Local session and logging storage
├── pkg/                 # Utility packages
├── static/              # Frontend assets
└── scripts/             # External tool scripts callable by LLM agent
```

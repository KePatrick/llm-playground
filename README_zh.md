# LLMplayground(go-clean-architecture)
## 專案概覽

### 前言

這是一個於閒暇時間開發的練習型專案，旨在學習與實作以下目標：

- 熟悉 **Go 語言**
- 掌握 **Clean Architecture（清潔架構）** 的實踐方式
- 理解與實作 **OpenAI 相容 API function call 流程**
- 實作一個支援記憶與工具呼叫的聊天助手

本專案聚焦在結構清晰與功能正確性，僅作為練習與參考用途。如有問題或建議，歡迎交流。

---

## 實作功能

- 流式回應（Streaming Response）
- 對話記憶儲存（Redis，可選，預設存於local）
- 輸出輸入紀錄（MySQL，可選，預設存於local）
- 工具呼叫（Function Calling）
- JSON 設定檔自訂模型與啟用模組

注意: 目前openai api 流式回應選擇顯示token用量時不知為何回覆異常，所以如選擇openai api 則token用量會無法紀錄

---

## Clean Architecture 分層說明

本專案遵循 Clean Architecture 分層設計，結構清晰，便於測試與維護。分層概念如下：

| 層級 | 對應資料夾 | 職責與說明 |
|------|-------------|------------|
| **Domain（領域層）** | `internal/domain/entity`, `internal/domain/repository` | 定義商業實體與邏輯契約，不依賴任何外部框架 |
| **Use Case（應用層）** | `internal/usecase` | 實作應用邏輯（如聊天流程、工具選擇），調用 domain 物件 |
| **Interface Adapters（介面轉換層）** | `internal/gateway/http` | 控制器邏輯，處理 HTTP 請求與回應格式轉換 |
| **Frameworks & Drivers（外部框架）** | `internal/infra`, `scripts/` | 實作資料庫、Redis、API 等實際依賴邏輯 |
| **進入點與設定** | `cmd/app`, `configs/` | 初始化應用、讀取設定與啟動服務 |
| **共用工具與模組** | `pkg/` | 非核心邏輯的工具套件 |


**依賴規則（Dependency Rule）**
   - **原則**：程式碼依賴方向只能由外層指向內層，內層不應知道外層的存在。
     - 例如：HTTP 控制器 (外層) 可以呼叫 Use Case (內層)，但 Use Case 不可直接 import HTTP 控制器。
   - **技術實作**：
     - 透過介面 (Interface) 定義契約。內層僅依賴這些介面，而介面的實作放在外層。
     - 在外層 (如 `main.go` 或 DI 容器) 進行依賴注入 (Dependency Injection)，將具體實作注入內層的介面。
   - **好處**：
     - **解耦合**：內層不受外層框架、庫、工具變化影響。
     - **易於替換**：可在不改動核心邏輯的情況下替換技術棧 (如更換資料庫、網路庫)。
     - **可測試性**：在測試時可以用 Mock 或 Fake 取代外層實作，專注測試內層邏輯。
---

## 系統需求

- OpenAI 相容 API 金鑰 (模型需支援function call)
- 自訂工具腳本
- MySQL（選用）
- Redis（選用）

---

## 前置設定

### 1. 下載專案

```bash
git clone https://github.com/yourusername/your-repo.git
cd your-repo
```

### 2. 設定 config 檔案

#### `configs/options.json`（必填）

```json
{
  "selectApi": "deepseek-chat", 
  "sysPrompt": "you are a assistant about this project: llmplayground", 
  "relationDatabase": false, 
  "redis": false 
}
```
註記: 
 - selectApi: 對應到configs/api.json
 - sysPrompt: 系統提示詞
 - relationDatabase: 是否啟用DB(僅支援MySQL)
 - redis: 是否啟用redis

> `relationDatabase` 與 `redis` 預設為 false，如設為 true，需額外設定`configs/database.json`, `configs/redis.json`。

#### `configs/api.json`（必填）

```json
{
  "deepseek-chat": {
    "apiKey": "your-api-key",
    "model": "deepseek-chat",
    "apiUrl": "https://api.deepseek.com/chat/completions"
  }
}
```

#### `configs/tools.json`（可選）

```json
[
  {
    "type": "function",
    "function": {
      "name": "fetchProjectInfo",
      "description": "查詢本專案資訊",
      "parameters": {
        "type": "object",
        "properties": {
          "languege": {
            "type": "string",
            "description": "使用者語言，選項: zh, en"
          }
        },
        "required": ["languege"]
      }
    },
    "script": "fetchProjectInfo"
  }
]
```

#### `configs/database.json`（使用 MySQL 時填寫）

```json
{
  "driver": "mysql",
  "name": "myDb",
  "url": "localhost:3306",
  "user": "user",
  "password": "password"
}
```
註記: 
 - driver: 目前僅支援MySQL

#### `configs/redis.json`（使用 Redis 時填寫）

```json
{
  "addr": "localhost:6379",
  "password": "",
  "db": 0
}
```

---

## 編譯與執行

```bash
go build -o app ./cmd/app
./app
```

- 預設 HTTP 伺服器監聽於 `localhost:8080`

---

## 使用方式

- 開啟瀏覽器連至 `localhost:8080` 進行對話
- 聊天室編號存於瀏覽器session（關閉視窗後記憶將清除）

---

## 專案結構摘要

```
├── cmd/app              # 應用程式主入口 (Go 原始碼)
├── configs/             # 系統設定檔（JSON 格式）
├── internal/
│   ├── config/          # 設定解析 (Configuration Parser)
│   ├── domain/          # 商業邏輯模型 (Entity)、倉儲與服務介面 (Repository & Service Interface)
│   │   ├── entity/
│   │   ├── repository/
│   │   └── service/
│   ├── usecase/         # 應用服務流程 (Use Case)
│   ├── gateway/http/    # HTTP 控制器與 DTO (Interface Adapters)
│   └── infra/           # 外部框架與驅動 (Frameworks & Drivers)：DB、Redis、LLM、Local
├── local/               # 本機儲存 (Record & Session)，在未啟用 MySQL/Redis 時使用
├── pkg/                 # 共用工具與模組 (Utility Packages)
├── static/              # 前端靜態資源
└── scripts/             # 可被 LLM Agent 呼叫的外部工具腳本
```


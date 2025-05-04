package config

import (
	"kepatrick/llm-playground/internal/config/reader"
	"log"
)

type DbConfig struct {
	Driver   string
	Name     string
	Url      string
	User     string
	Password string
}

type Apis map[string]ApiConfig

type ApiConfig struct {
	Model  string `json:"model"`
	ApiKey string `json:"apiKey"`
	ApiUrl string `json:"apiUrl"`
}

type Option struct {
	SelectApi        string `json:"selectApi"`
	SysPrompt        string `json:"sysPrompt"`
	RelationDatabase bool   `json:"relationDatabase"`
	Redis            bool   `json:"redis"`
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
	Script   string   `json:"script,omitempty"`
}

// Function represents the function object within a tool
type Function struct {
	Description string     `json:"description"`
	Name        string     `json:"name"`
	Parameters  Parameters `json:"parameters"`
}

// Parameters represents the parameters object within a function
type Parameters struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required"`
}

type Redis struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	Db       int    `json:"db"`
}

var (
	Options Option
	Tools   []Tool
)

func init() {
	Options = LoadOption()
}

func LoadDbConfig() DbConfig {
	rslt, err := reader.LoadJsonConfig[DbConfig]("./configs/database.json")
	if err != nil {
		log.Fatalf("fail to load database config, err: %v", err)
	}
	return rslt
}

func LoadLlmConfig() ApiConfig {
	apis, err := reader.LoadJsonConfig[Apis]("./configs/api.json")
	if err != nil {
		log.Fatalf("fail to load api config, err: %v", err)
	}

	return apis[Options.SelectApi]
}

func LoadOption() Option {
	options, err := reader.LoadJsonConfig[Option]("./configs/options.json")
	if err != nil {
		log.Fatalf("fail to load options, err: %v", err)
	}
	return options
}

func LoadToolDef() []Tool {
	tools, err := reader.LoadJsonConfig[[]Tool]("./configs/tools.json")
	if err != nil {
		log.Fatalf("fail to load tools, err: %v", err)
	}

	return tools
}

func LoadRedis() Redis {
	redis, err := reader.LoadJsonConfig[Redis]("./configs/redis.json")
	if err != nil {
		log.Fatalf("fail to load redis config, err: %v", err)
	}
	return redis
}

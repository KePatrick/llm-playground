package main

import (
	"kepatrick/llm-playground/internal/config"
	"kepatrick/llm-playground/internal/domain/repository"
	httpAdapter "kepatrick/llm-playground/internal/gateway/http"

	"kepatrick/llm-playground/internal/infra/database"
	"kepatrick/llm-playground/internal/infra/llm"
	"kepatrick/llm-playground/internal/infra/local"
	"kepatrick/llm-playground/internal/infra/redis"

	"kepatrick/llm-playground/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadLlmConfig()
	// Infra init
	httpClient := &http.Client{}

	logRepo := getLogRepo(config.LoadOption())
	sessRepo := getSessionRepo(config.LoadOption())

	// init llm
	llmSvc := llm.NewOpenAILLMService(cfg.ApiKey, cfg.ApiUrl, cfg.Model, httpClient, config.LoadToolDef())

	// Usecase init
	genUsecase := usecase.NewGenerateUsecase(llmSvc, sessRepo, logRepo)

	// HTTP Server
	r := gin.Default()
	httpAdapter.RegisterRoutes(r, genUsecase)

	r.Run(":8080")
}

func getSessionRepo(cfg config.Option) repository.SessionRepository {
	var sessionRepo repository.SessionRepository
	if cfg.Redis {
		// init redis
		redisClient := redis.InitRedisClient(config.LoadRedis())
		sessionRepo = redis.NewRedisSessionRepo(redisClient)
	} else {
		sessionRepo = local.NewFileSessionRepo("./local/session/")
	}
	return sessionRepo
}

func getLogRepo(cfg config.Option) repository.LogRepository {
	var logRepo repository.LogRepository
	//init database
	if cfg.RelationDatabase {
		if config.LoadDbConfig().Driver == "mysql" {
			db := database.InitDb(config.LoadDbConfig())
			logRepo = database.NewLogRepository(db)
		}
	} else {
		logRepo = local.NewExcelLogRepo("./local/record/record.xlsx")
	}
	return logRepo
}

package database

import (
	"fmt"
	"kepatrick/llm-playground/internal/config"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type LogRepository struct {
	dbClient *gorm.DB
}

func InitDb(dbCnf config.DbConfig) *gorm.DB {
	dsn := buildConnectionString(dbCnf)

	var err error
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("fialed to connet to database: " + err.Error())
	}

	return db
}

func buildConnectionString(dbCnf config.DbConfig) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbCnf.User, dbCnf.Password, dbCnf.Url, dbCnf.Name,
	)
}

func NewLogRepository(db *gorm.DB) *LogRepository {
	return &LogRepository{
		db,
	}
}

func (r *LogRepository) Insert(sessionId, reqMsg, resMsg string, reqToken, resToken int, sendTime, receiveTime time.Time) error {

	r.dbClient.Create(&Record{
		Id:          time.Now().Format("20060102150405"),
		ChatId:      sessionId,
		ReqMessage:  reqMsg,
		ResMessage:  resMsg,
		Prompt:      "",
		ReqToken:    reqToken,
		ResToken:    resToken,
		SendTime:    sendTime,
		ReceiveTime: receiveTime,
	})
	return nil
}

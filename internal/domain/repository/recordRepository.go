package repository

import "time"

type LogRepository interface {
	Insert(sessionId, reqMsg, resMsg string, reqToken, resToken int, sendTime, receiveTime time.Time) error
}

package database

import "time"

type Record struct {
	Id          string
	ChatId      string
	ReqMessage  string
	ResMessage  string
	Prompt      string
	ReqToken    int
	ResToken    int
	SendTime    time.Time
	ReceiveTime time.Time
}

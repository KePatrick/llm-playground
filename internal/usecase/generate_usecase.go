package usecase

import (
	"context"
	"fmt"
	"kepatrick/llm-playground/internal/config"
	"kepatrick/llm-playground/internal/domain/entity"
	"kepatrick/llm-playground/internal/domain/repository"
	"kepatrick/llm-playground/internal/domain/service"
	"time"
)

type GenerateUsecase struct {
	llmSvc      service.LLMService
	sessionRepo repository.SessionRepository
	logRepo     repository.LogRepository
}

func NewGenerateUsecase(llmsvc service.LLMService, sessionRepo repository.SessionRepository, logRepo repository.LogRepository) *GenerateUsecase {
	return &GenerateUsecase{llmsvc, sessionRepo, logRepo}
}

func (u *GenerateUsecase) RunStream(ctx context.Context, sessionID, prompt string, writer service.StreamWriter) error {
	fmt.Printf("receive prompt:%s", prompt)
	if !u.sessionRepo.ExistKey(ctx, sessionID) {
		u.sessionRepo.AppendMessage(ctx, sessionID, entity.Message{"system", config.LoadOption().SysPrompt, "", nil, nowMilli()})
	}
	sendTime := time.Now()
	u.sessionRepo.AppendMessage(ctx, sessionID, entity.Message{"user", prompt, "", nil, nowMilli()})
	messages, err := u.sessionRepo.FetchPrevMessage(ctx, sessionID)

	originMsgSize := len(messages)
	if err != nil {
		fmt.Printf("%v", err)
		return err
	}

	var llmRslt service.LLMResult
	callLLm := true

	for callLLm {
		llmRslt, err = u.llmSvc.StreamingCall(ctx, messages, writer, llmRslt)
		// update messages
		messages = llmRslt.Messages

		if err != nil {
			fmt.Printf("%v", err)
			return err
		}

		if !llmRslt.IsToolCall {
			callLLm = false
			break
		}
	}

	// Update session memory and Record
	go func() {
		for i := range llmRslt.Messages {
			if originMsgSize > i {
				continue
			}
			err := u.sessionRepo.AppendMessage(ctx, sessionID, llmRslt.Messages[i])
			if err != nil {
				fmt.Printf("error: %v", err)
			}
		}

		err := u.logRepo.Insert(sessionID, prompt, llmRslt.LlmRes, llmRslt.ReqToken, llmRslt.ResToken, sendTime, time.Now())

		if err != nil {
			fmt.Printf("error: %v", err)
		}
	}()

	// u.logRepo.Insert(ctx, sessionID, prompt)
	return nil
}

func nowMilli() string { return fmt.Sprintf("%d", time.Now().UnixMilli()) }

// // Session & Log interfaces define
//
// type SessionRepository interface {
//     AppendMessage(ctx context.Context, sessionID string, msg entity.Message) error
// 	FetchPrevMessage(ctx context.Context, sessionID string) ([]entity.Message, error)
// }
//
// type LogRepository interface {
//     Insert(sessionId, reqMsg, resMsg string, reqToken, resToken int, sendTime, receiveTime time.Time) error
// }

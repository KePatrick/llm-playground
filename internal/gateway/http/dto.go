package http

type GenerateRequest struct {
	SessionID string `json:"sessionId" binding:"required"`
	Prompt    string `json:"prompt" binding:"required"`
}

package http

import (
	"github.com/gin-gonic/gin"
)

// GinStreamWriter implements StreamWriter for Gin
type GinStreamWriter struct {
	c *gin.Context
}

func NewGinStreamWriter(c *gin.Context) *GinStreamWriter {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	return &GinStreamWriter{c}
}

func (w *GinStreamWriter) Write(data string) error {
	_, err := w.c.Writer.Write([]byte("data: " + data + "\n\n"))
	w.c.Writer.Flush()
	return err
}

func (w *GinStreamWriter) Done() error {
	_, err := w.c.Writer.Write([]byte("data: [DONE]\n\n"))
	w.c.Writer.Flush()
	return err
}

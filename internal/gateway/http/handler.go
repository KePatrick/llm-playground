package http

import (
	"html/template"
	usecase "kepatrick/llm-playground/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, u *usecase.GenerateUsecase) {

	r.Static("/static", "./static")
	// Set template
	r.SetHTMLTemplate(template.Must(template.ParseFiles("static/chat.html")))

	// Chat page routing
	r.GET("/chat", func(c *gin.Context) {
		c.HTML(http.StatusOK, "chat.html", gin.H{
			"APIUrl": "/generate",
		})
	})

	// Redirect to chat page
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/chat")
	})

	r.POST("/generate", func(c *gin.Context) {
		var req GenerateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// stream
		w := NewGinStreamWriter(c)
		if err := u.RunStream(c.Request.Context(), req.SessionID, req.Prompt, w); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	})
}

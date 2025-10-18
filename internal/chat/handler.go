package chat

import (
	"gochatty/internal/db"
	"gochatty/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type PostMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

func getUserID(username string) (int, error) {
	var id int
	err := db.DB.Get(&id, "SELECT id FROM users WHERE username=$1", username)
	return id, err
}

func PostMessage(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req PostMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content"})
		return
	}

	userID, err := getUserID(username.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	_, err = db.DB.Exec(`
		INSERT INTO messages (user_id, content, created_at, is_bot)
		VALUES ($1, $2, $3, false)
	`, userID, req.Content, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Message saved."})
}

func GetMessages(c *gin.Context) {
	var msgs []models.Message
	err := db.DB.Select(&msgs, `
		SELECT id, user_id, content, created_at, is_bot
		FROM messages
		ORDER BY created_at DESC
		LIMIT 50
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}

	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	c.JSON(http.StatusOK, msgs)
}

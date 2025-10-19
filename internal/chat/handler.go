package chat

import (
	"encoding/json"
	"gochatty/internal/db"
	"gochatty/internal/models"
	"gochatty/internal/mq"
	"gochatty/internal/websocket"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type PostMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

type StockCommand struct {
	UserID    int    `json:"user_id"`
	StockCode string `json:"stock_code"`
}

type BroadcastMessage struct {
	Content   string    `json:"content"`
	User      string    `json:"user"`
	Timestamp time.Time `json:"timestamp"`
}

var rabbitCmd, rabbitMsgs *mq.RabbitMQ

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

	if strings.HasPrefix(req.Content, "/stock=") {
		stockCode := strings.TrimPrefix(req.Content, "/stock=")

		cmd := StockCommand{
			UserID:    userID,
			StockCode: stockCode,
		}

		body, err := json.Marshal(cmd)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal cmd"})
			return
		}

		err = rabbitCmd.Publish(body)
		if err != nil {
			log.Printf("Failed to publish stock command: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue stock cmd"})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{"message": "Stock command received and processing"})
		return
	}

	msg := models.Message{
		UserID:    userID,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	err = SaveMessage(msg)
	if err != nil {
		log.Printf("Failed to save message: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
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

func InitRabbitMQClient(cmd *mq.RabbitMQ, msgs *mq.RabbitMQ) {
	rabbitCmd = cmd
	rabbitMsgs = msgs
}

func SaveMessage(message models.Message) error {
	_, err := db.DB.Exec(`
		INSERT INTO messages (user_id, content, created_at, is_bot)
		VALUES ($1, $2, $3, false)
	`, message.UserID, message.Content, message.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

func BroadcastMessageToClients(content, user string, timestamp time.Time) {
	msg := BroadcastMessage{
		Content:   content,
		User:      user,
		Timestamp: timestamp,
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal broadcast message: %v", err)
	}

	websocket.Broadcast <- msgBytes
}

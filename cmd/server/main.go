package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"gochatty/internal/auth"
	"gochatty/internal/chat"
	"gochatty/internal/db"
	"gochatty/internal/models"
	"gochatty/internal/mq"
	"gochatty/internal/websocket"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const BOT_ID = 1

type StockCommand struct {
	UserID    uint   `json:"user_id"`
	StockCode string `json:"stock_code"`
}

func fetchStockQuote(stockCode string) (string, error) {
	url := fmt.Sprintf("https://stooq.com/q/l/?s=%s&f=sd2t2ohlcv&h&e=csv", stockCode)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	reader := csv.NewReader(bufio.NewReader(resp.Body))
	_, err = reader.Read()
	if err != nil {
		return "", err
	}

	record, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return "", fmt.Errorf("stock quote not found")
		}
		return "", err
	}

	closePrice := record[6]
	if closePrice == "N/D" {
		return "", fmt.Errorf("no data for stock %s", stockCode)
	}

	return fmt.Sprintf("%s quote is $%s per share", strings.ToUpper(stockCode), closePrice), nil
}

func runBot(cmdClient *mq.RabbitMQ, msgClient *mq.RabbitMQ) {
	msgs, err := cmdClient.Consume(false)
	if err != nil {
		log.Fatalf("Bot: Failed to consume command queue: %v", err)
	}
	log.Println("Bot: Started consuming stock_commands")

	for d := range msgs {
		var cmd StockCommand
		err := json.Unmarshal(d.Body, &cmd)
		if err != nil {
			log.Printf("Bot: Invalid command message: %v", err)
			d.Ack(false)
			continue
		}

		quote, err := fetchStockQuote(cmd.StockCode)
		if err != nil {
			quote = fmt.Sprintf("Error fetching quote for %s: %v", cmd.StockCode, err)
		}

		botMessage := chat.PostMessageRequest{
			Content: quote,
		}

		body, err := json.Marshal(botMessage)
		if err != nil {
			log.Printf("Bot: Failed to marshal botMessage: %v", err)
			d.Ack(false)
			continue
		}

		err = msgClient.Publish(body)
		if err != nil {
			log.Printf("Bot: Failed to publish botMessage: %v", err)
			d.Ack(false)
			continue
		}

		d.Ack(false)
		log.Println("Bot: Message acknowledged")
	}
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	db.Init(dsn)

	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@localhost:5672/"
	}

	rabbitCmd, err := mq.NewRabbitMQ(rabbitMQURL, "stock_commands")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ for commands: %v", err)
	}
	defer rabbitCmd.Close()

	rabbitMsgs, err := mq.NewRabbitMQ(rabbitMQURL, "chat_messages")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ for messages: %v", err)
	}
	defer rabbitMsgs.Close()

	chat.InitRabbitMQClient(rabbitCmd, rabbitMsgs)

	go runBot(rabbitCmd, rabbitMsgs)

	r := gin.Default()

	r.POST("/register", auth.Register)
	r.POST("/login", auth.Login)

	authorized := r.Group("/")
	authorized.Use(auth.JWTAuthMiddleware())

	authorized.POST("/messages", chat.PostMessage)
	authorized.GET("/messages", chat.GetMessages)

	go websocket.HandleMessages()
	r.GET("/ws", websocket.HandleConnections)

	go func() {
		msgs, err := rabbitMsgs.Consume(false)
		if err != nil {
			log.Fatalf("Failed to consume chat messages: %v", err)
		}

		log.Println("Started consuming chat_messages")

		for d := range msgs {
			var msg chat.PostMessageRequest
			err := json.Unmarshal(d.Body, &msg)
			if err != nil {
				log.Printf("Failed to unmarshal chat post message: %v", err)
				d.Ack(false)
				continue
			}

			message := models.Message{
				UserID:    BOT_ID,
				Content:   msg.Content,
				CreatedAt: time.Now(),
			}

			log.Println("Saving bot message")
			chat.SaveMessage(message)

			chat.BroadcastMessageToClients(msg.Content, "bot", BOT_ID, message.CreatedAt)
			d.Ack(false)
		}
	}()

	log.Println("Server running on port 8080")
	log.Fatal(r.Run(":8080"))
}

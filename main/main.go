package main

import (
	"gochatty/internal/auth"
	"gochatty/internal/chat"
	"gochatty/internal/db"
	"gochatty/internal/websocket"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	db.Init(dsn)

	r := gin.Default()
	r.POST("/register", auth.Register)
	r.POST("/login", auth.Login)

	authorized := r.Group("/")
	authorized.Use(auth.JWTAuthMiddleware())

	authorized.POST("/messages", chat.PostMessage)
	authorized.GET("/messages", chat.GetMessages)

	go websocket.HandleMessages()
	r.GET("/ws", websocket.HandleConnections)

	log.Fatal(r.Run(":8080"))
}

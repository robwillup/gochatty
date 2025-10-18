package main

import (
	"gochatty/internal/auth"
	"gochatty/internal/db"
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

	log.Fatal(r.Run(":8080"))
}

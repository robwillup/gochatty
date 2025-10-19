package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Client struct {
	Conn   *websocket.Conn
	Send   chan BroadcastMessage
	UserID int
}

type BroadcastMessage struct {
	Content   string    `json:"content"`
	User      string    `json:"user"`
	UserID    int       `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	clients   = make(map[*Client]bool)
	Broadcast = make(chan BroadcastMessage)
	mutex     = &sync.Mutex{}
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func HandleConnections(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket Upgrade: %v", err)
		return
	}
	defer ws.Close()

	client := &Client{Conn: ws, Send: make(chan BroadcastMessage), UserID: -1}
	mutex.Lock()
	clients[client] = true
	mutex.Unlock()

	go func() {
		for msg := range client.Send {
			msgBytes, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Failed to marshal broadcast message: %v", err)
			}
			err = client.Conn.WriteMessage(websocket.TextMessage, msgBytes)
			if err != nil {
				log.Printf("Write error: %v", err)
				break
			}
		}
	}()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			mutex.Lock()
			delete(clients, client)
			mutex.Unlock()
			break
		}
		var broadcastMessage BroadcastMessage
		err = json.Unmarshal(msg, &broadcastMessage)
		if err != nil {
			log.Printf("Failed to unmarshal broadcast message: %v", err)
		}
		Broadcast <- broadcastMessage
	}
}

func HandleMessages() {
	for {
		msg := <-Broadcast
		mutex.Lock()
		for client := range clients {
			if client.UserID == msg.UserID {
				continue
			}
			select {
			case client.Send <- msg:
			default:
				close(client.Send)
				delete(clients, client)
			}
		}
		mutex.Unlock()
	}
}

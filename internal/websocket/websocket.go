package websocket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Client struct {
	Conn *websocket.Conn
	Send chan []byte
}

var (
	clients   = make(map[*Client]bool)
	broadcast = make(chan []byte)
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

	client := &Client{Conn: ws, Send: make(chan []byte)}
	mutex.Lock()
	clients[client] = true
	mutex.Unlock()

	go func() {
		for msg := range client.Send {
			err := client.Conn.WriteMessage(websocket.TextMessage, msg)
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
		broadcast <- msg
	}
}

func HandleMessages() {
	for {
		msg := <-broadcast
		mutex.Lock()
		for client := range clients {
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

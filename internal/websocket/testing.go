package websocket

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

var clients = make(map[*websocket.Conn]bool)

var broadcast = make(chan string)

func HandleTestingConnections(c *gin.Context) {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("upgrade gagal:", err)
		return
	}
	defer ws.Close()

	clients[ws] = true
	log.Println("client baru terhubung, total:", len(clients))

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("client terputus:", err)
			delete(clients, ws)
			break
		}
		broadcast <- string(msg)
	}
}

func HandleTestingBroadcast() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Println("gagal kirim ke client:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

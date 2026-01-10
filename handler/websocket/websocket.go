package websocket

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleConnections(c *gin.Context) {
	// 升级HTTP连接为WebSocket
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}
	defer ws.Close()

	// 循环处理消息
	for {
		var msg string
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		log.Printf("Received: %s", msg)

		// 回传消息
		reply := fmt.Sprintf("Server received: %s", msg)
		err = ws.WriteJSON(reply)
		if err != nil {
			log.Printf("Write error: %v", err)
			break
		}
	}
}

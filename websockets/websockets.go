package websockets

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WebsocketHandler(w http.ResponseWriter, r *http.Request, processing chan<- []byte) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Client connected!")

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		processing <- message

		msg := fmt.Sprintf("Event sent for processing.")

		err = conn.WriteMessage(messageType, []byte(msg))
		if err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}

package websockets

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Event struct {
	Data json.RawMessage `json:"data"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WebsocketHandler(w http.ResponseWriter, r *http.Request, processing chan<- map[string]any) {
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

		var event Event
		if err := json.Unmarshal(message, &event); err != nil {
			log.Println("Invalid JSON received:", err)
			conn.WriteMessage(messageType, []byte("Error: Invalid JSON format\n"))
			continue
		}

		payload := map[string]any{
			"data": event.Data,
		}

		processing <- payload

		msg := fmt.Sprintf("Event acknowledged: %s\n", event.Data)

		err = conn.WriteMessage(messageType, []byte(msg))
		if err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}

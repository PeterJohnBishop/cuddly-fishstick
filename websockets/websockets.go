package websockets

import (
	"encoding/json"
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

func WebsocketHandler(w http.ResponseWriter, r *http.Request, processing chan<- map[string]any) {

	// Expects:
	// {
	// 	"event": string,
	// 	"data":  string,
	// }

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

		var data map[string]interface{}
		err = json.Unmarshal(message, &data)
		if err != nil {
			log.Println("Failed to unmarshal message")
		}

		if eventVal, exists := data["event"]; exists {
			if eventString, ok := eventVal.(string); ok {
				validEvent := map[string]any{
					"event": eventString,
					"data":  data["data"],
				}
				processing <- validEvent
			} else {
				log.Println("Property 'event' found, but it is not a string")
			}
		} else {
			log.Println("Property 'event' is missing from the payload")
		}

		msg := "Event sent for processing."

		err = conn.WriteMessage(messageType, []byte(msg))
		if err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}

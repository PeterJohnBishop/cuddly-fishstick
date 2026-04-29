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

		var msgToClient string

		if eventVal, exists := data["event"]; exists {
			if eventString, ok := eventVal.(string); ok {
				validEvent := map[string]any{
					"event": eventString,
					"data":  data["data"],
				}

				select {
				case processing <- validEvent:
					msgToClient = `{"status": "success", "message": "Event sent for processing."}`
				default:
					log.Println("Server at capacity: dropped valid WS event")
					msgToClient = `{"status": "error", "message": "Server is at capacity. Please retry later."}`
				}

			} else {
				log.Println("Property 'event' found, but it is not a string")
				msgToClient = `{"status": "error", "message": "Invalid event format."}`
			}
		} else {
			log.Println("Property 'event' is missing from the payload")
			unknownEvent := map[string]any{
				"event": "unknown",
				"data":  data,
			}

			select {
			case processing <- unknownEvent:
				msgToClient = `{"status": "success", "message": "Event sent for processing."}`
			default:
				log.Println("Server at capacity: dropped unknown WS event")
				msgToClient = `{"status": "error", "message": "Server is at capacity. Please retry later."}`
			}
		}

		err = conn.WriteMessage(messageType, []byte(msgToClient))
		if err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}

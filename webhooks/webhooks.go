package webhooks

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

func sendJSONResponse(w http.ResponseWriter, statusCode int, status, data string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]string{
		"data": data,
	}
	json.NewEncoder(w).Encode(response)
}

func WebhookHandler(w http.ResponseWriter, r *http.Request, processing chan<- map[string]any) {

	// checks if 'event' is either query param or in the body itself

	// Expects:
	// {
	// 	"event": string,
	// 	"data":  string,
	// }

	// if no event, "event": "unknown"

	queryParams := r.URL.Query()
	eventParam := queryParams.Get("event")

	if r.Method != http.MethodPost {
		sendJSONResponse(w, http.StatusMethodNotAllowed, "error", "Method Not Allowed")
		return
	}

	const maxPayloadSize = 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxPayloadSize)

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			sendJSONResponse(w, http.StatusRequestEntityTooLarge, "error", "Payload too large. Maximum size is 1MB.")
			return
		}

		sendJSONResponse(w, http.StatusBadRequest, "error", "Failed to read request body")
		return
	}

	if len(body) == 0 {
		sendJSONResponse(w, http.StatusBadRequest, "error", "Missing required payload")
		return
	}

	if eventParam == "" {
		var data map[string]interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			sendJSONResponse(w, http.StatusBadRequest, "error", "Unable to unmarshal the payload")
		}

		if eventVal, exists := data["event"]; exists {
			if eventString, ok := eventVal.(string); ok {
				validEvent := map[string]any{
					"event": eventString,
					"data":  data["data"],
				}
				select {
				case processing <- validEvent:
					sendJSONResponse(w, http.StatusOK, "success", "Webhook received and queued successfully")

				default:
					sendJSONResponse(w, http.StatusTooManyRequests, "error", "Server is at capacity. Please retry later.")
				}
			} else {
				log.Println("Property 'event' found, but it is not a string")
			}
		} else {
			log.Println("Property 'event' is missing from the payload")
			unknownEvent := map[string]any{
				"event": "unknown",
				"data":  data,
			}
			select {
			case processing <- unknownEvent:
				sendJSONResponse(w, http.StatusOK, "success", "Webhook received and queued successfully")
			default:
				sendJSONResponse(w, http.StatusTooManyRequests, "error", "Server is at capacity. Please retry later.")
			}
		}
	} else {
		validEvent := map[string]any{
			"event": eventParam,
			"data":  body,
		}

		select {
		case processing <- validEvent:
			sendJSONResponse(w, http.StatusOK, "success", "Webhook received and queued successfully")

		default:
			sendJSONResponse(w, http.StatusTooManyRequests, "error", "Server is at capacity. Please retry later.")
		}
	}

}

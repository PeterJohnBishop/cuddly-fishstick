package webhooks

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Event struct {
	Data json.RawMessage `json:"data"`
}

func sendJSONResponse(w http.ResponseWriter, statusCode int, status, data string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]string{
		"data": data,
	}
	json.NewEncoder(w).Encode(response)
}

func WebhookHandler(w http.ResponseWriter, r *http.Request, processing chan<- map[string]any) {

	// Accept only POST requests
	if r.Method != http.MethodPost {
		sendJSONResponse(w, http.StatusMethodNotAllowed, "error", "Method Not Allowed")
		return
	}

	// Limit to 1MB
	const maxPayloadSize = 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxPayloadSize)

	defer r.Body.Close()

	var event Event
	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			sendJSONResponse(w, http.StatusRequestEntityTooLarge, "error", "Payload too large. Maximum size is 1MB.")
			return
		}

		sendJSONResponse(w, http.StatusBadRequest, "error", "Invalid JSON payload structure")
		return
	}

	// validate that the event field isn't empty
	if string(event.Data) == "" {
		sendJSONResponse(w, http.StatusBadRequest, "error", "Missing required 'event' field")
		return
	}

	payload := map[string]any{
		"event": event.Data,
	}

	// Send the payload for processing
	select {
	case processing <- payload:
		// Send a success response immediately
		sendJSONResponse(w, http.StatusOK, "success", "Webhook received and queued successfully")

	default:
		// The channel is completely full!
		// Sending 429 error to let the sender know they can retry later.
		sendJSONResponse(w, http.StatusTooManyRequests, "error", "Server is at capacity. Please retry later.")
	}
}

package webhooks

import (
	"encoding/json"
	"errors"
	"net/http"
)

type EventWrapper struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

func sendJSONResponse(w http.ResponseWriter, statusCode int, status, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]string{
		"event": status,
		"data":  message,
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

	var wrapper EventWrapper
	err := json.NewDecoder(r.Body).Decode(&wrapper)
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
	if wrapper.Event == "" {
		sendJSONResponse(w, http.StatusBadRequest, "error", "Missing required 'event' field")
		return
	}

	payload := map[string]any{
		"event": wrapper.Event,
		"data":  wrapper.Data,
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

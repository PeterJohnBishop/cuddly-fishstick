package webhooks

import (
	"encoding/json"
	"errors"
	"net/http"
)

func sendJSONResponse(w http.ResponseWriter, statusCode int, status, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]string{
		"status":  status,
		"message": message,
	}
	json.NewEncoder(w).Encode(response)
}

func GeneralWebhookHandler(w http.ResponseWriter, r *http.Request, processing chan<- map[string]any) {

	var payload map[string]any

	// accept only POST requests
	if r.Method != http.MethodPost {
		sendJSONResponse(w, http.StatusMethodNotAllowed, "error", "Method Not Allowed")
		return
	}

	// Limit to 1MB
	const maxPayloadSize = 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxPayloadSize)

	defer r.Body.Close()

	// decode the payload to JSON
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		// check if the error was caused by exceeding the MaxBytes limit
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			sendJSONResponse(w, http.StatusRequestEntityTooLarge, "error", "Payload too large. Maximum size is 1MB.")
			return
		}

		sendJSONResponse(w, http.StatusBadRequest, "error", "Invalid JSON payload")
		return
	}

	// send the payload for processing
	processing <- payload

	// send a success response immediately
	sendJSONResponse(w, http.StatusOK, "success", "Webhook received and queued")
}

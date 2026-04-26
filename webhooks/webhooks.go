package webhooks

import (
	"encoding/json"
	"errors"
	"io"
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

func WebhookHandler(w http.ResponseWriter, r *http.Request, processing chan<- []byte) {

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

	select {
	case processing <- body:
		sendJSONResponse(w, http.StatusOK, "success", "Webhook received and queued successfully")

	default:
		sendJSONResponse(w, http.StatusTooManyRequests, "error", "Server is at capacity. Please retry later.")
	}
}

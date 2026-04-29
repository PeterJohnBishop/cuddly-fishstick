package server

import (
	"cuddly-fishstick/webhooks"
	"cuddly-fishstick/websockets"
	"encoding/json"
	"net/http"
)

func AddRoutes(processing chan<- map[string]any) *http.ServeMux {
	mux := http.NewServeMux()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		data := map[string]string{"message": "You've reached the Cuddly-Fishstick, please leave a message after the "}
		json.NewEncoder(w).Encode(data)
	})

	mux.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		webhooks.WebhookHandler(w, r, processing)
	})

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websockets.WebsocketHandler(w, r, processing)
	})

	return mux
}

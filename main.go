package main

import (
	"cuddly-fishstick/webhooks"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	baseURL := os.Getenv("BASE_URL")
	port := os.Getenv("PORT")
	processing := make(chan map[string]any)

	go func() {
		log.Println("Worker started, waiting for data...")
		for payload := range processing {
			fmt.Printf("[PROCESSED] %s\n", payload)
		}
	}()

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		webhooks.GeneralWebhookHandler(w, r, processing)
	})

	connstring := fmt.Sprintf("Server listing at %s on port %s", baseURL, port)
	log.Println(connstring)
	log.Fatal(http.ListenAndServe(port, nil))
}

package main

import (
	"context"
	"cuddly-fishstick/webhooks"
	"cuddly-fishstick/websockets"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found; relying on system environment variables.")
	}

	baseURL := os.Getenv("BASE_URL")
	port := os.Getenv("PORT")

	queueSize := 100
	processing := make(chan map[string]any, queueSize)

	var wg sync.WaitGroup

	// run processing in a waitgroup to track when processing is complete
	wg.Go(func() {
		log.Println("Worker started, waiting for data...")

		for payload := range processing {
			fmt.Printf("[PROCESSED] %v\n", payload)
			// time.Sleep(1 * time.Second) // Uncomment to test the graceful shutdown
		}
		log.Println("Queue drained. Worker shutting down.")
	})

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		webhooks.WebhookHandler(w, r, processing)
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websockets.WebsocketHandler(w, r, processing)
	})

	srv := &http.Server{
		Addr: ":" + port,
	}

	// separate goroutine so the server doesn't block main()
	go func() {
		connstring := fmt.Sprintf("Server listening at %s on port %s", baseURL, port)
		log.Println(connstring)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 5listen for system interrupt signals
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// user kills with CTRL-C
	// Docker sends SIGTERM on 'docker stop'

	// block main() here until an interrupt signal is received
	<-quit
	log.Println("Shutdown signal received...")

	// This gives active HTTP handlers 5 seconds to finish replying to webhooks
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Shutting down HTTP server...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}

	log.Println("Closing processing channel...")
	close(processing)

	log.Println("Waiting for worker to drain the remaining queue...")
	wg.Wait()

	log.Println("Server exited cleanly.")
}

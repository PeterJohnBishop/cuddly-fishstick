package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func InitServer() {
	baseURL := os.Getenv("BASE_URL")
	port := os.Getenv("PORT")

	queueSize := 100
	processing := make(chan map[string]any, queueSize)
	var wg sync.WaitGroup

	// FIX: Standard sync.WaitGroup requires Add(1), go func(), and defer Done()
	wg.Add(1)
	go func() {
		defer wg.Done() // Ensure WaitGroup counter decrements when this exits
		log.Println("Worker started, waiting for data...")

		for payload := range processing {
			event := payload["event"]

			switch event {
			case "send_email":
				// send data to mail channel
			default:

			}

			log.Printf("Processing event: %s", event)
			// time.Sleep(1 * time.Second) // Uncomment to test the graceful shutdown
		}
		log.Println("Queue drained. Worker shutting down.")
	}()

	mux := AddRoutes(processing)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// separate goroutine so the server doesn't block main()
	go func() {
		connstring := fmt.Sprintf("Server listening at %s on port %s", baseURL, port)
		log.Println(connstring)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// listen for system interrupt signals
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

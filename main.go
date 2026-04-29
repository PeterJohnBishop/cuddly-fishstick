package main

import (
	"cuddly-fishstick/server"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found; relying on system environment variables.")
	}

	server.InitServer()
}

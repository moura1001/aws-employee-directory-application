package main

import (
	"log"
	"net/http"

	server "github.com/moura1001/aws-employee-directory-application/server/handler"
)

func main() {
	log.Println("Attempting to start server on port 5000...")

	server, err := server.NewServer()
	if err != nil {
		log.Fatalf("server startup error: %v", err)
	}

	if err := http.ListenAndServe(":5000", server); err != nil {
		log.Fatalf("could not listen on port 5000: %v", err)
	}
}

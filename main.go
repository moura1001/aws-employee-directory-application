package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	server "github.com/moura1001/aws-employee-directory-application/server/handler"
	"github.com/moura1001/aws-employee-directory-application/server/utils"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}

	PHOTOS_BUCKET := os.Getenv("PHOTOS_BUCKET")
	CSRF_SECRET := os.Getenv("CSRF_SECRET")
	DATABASE_HOST := os.Getenv("DATABASE_HOST")
	DATABASE_USER := os.Getenv("DATABASE_USER")
	DATABASE_PASSWORD := os.Getenv("DATABASE_PASSWORD")
	DATABASE_DB_NAME := os.Getenv("DATABASE_DB_NAME")
	DYNAMO_MODE := os.Getenv("DYNAMO_MODE")

	utils.PHOTOS_BUCKET = PHOTOS_BUCKET
	utils.CSRF_SECRET = CSRF_SECRET
	utils.DATABASE_HOST = DATABASE_HOST
	utils.DATABASE_USER = DATABASE_USER
	utils.DATABASE_PASSWORD = DATABASE_PASSWORD
	utils.DATABASE_DB_NAME = DATABASE_DB_NAME
	utils.DYNAMO_MODE = DYNAMO_MODE

	/*utils.PHOTOS_BUCKET = os.Getenv("PHOTOS_BUCKET")
	utils.CSRF_SECRET = os.Getenv("CSRF_SECRET")

	utils.DATABASE_HOST = os.Getenv("DATABASE_HOST")
	utils.DATABASE_USER = os.Getenv("DATABASE_USER")
	utils.DATABASE_PASSWORD = os.Getenv("DATABASE_PASSWORD")
	utils.DATABASE_DB_NAME = os.Getenv("DATABASE_DB_NAME")

	utils.DYNAMO_MODE = os.Getenv("DYNAMO_MODE")*/
}

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

package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/HamstimusPrime/text-analyzer-api/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

//based on the schema of the database and this text, generate a query that best fulfills this request,
//and if you can't, return an error code.

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	//establish DB connection
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("unable to establish connection to database: %v", err)
	}

	dbQueries := database.New(db)
	apiConfiguration := apiConfig{
		DB: dbQueries,
	}

	//setup a server

	port := os.Getenv("PORT")
	mux := http.NewServeMux()
	mux.HandleFunc("POST /strings", apiConfiguration.CreateText)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("server running on port: %v\n", port)
	log.Fatal(server.ListenAndServe())
}

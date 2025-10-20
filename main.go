package main

import (
	"log"
	"net/http"
	"os"

	"github.com/HamstimusPrime/text-analyzer-api/handlers"
)

//based on the schema of the database and this text, generate a query that best fulfills this request,
//and if you can't, return an error code.

func main() {
	//setup a server

	port := os.Getenv("PORT")
	mux := http.NewServeMux()
	mux.HandleFunc("POST /strings", handlers.CreateText)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("server running on port: %v\n", port)
	log.Fatal(server.ListenAndServe())
}

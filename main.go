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

	//setup Maps of accepted queries
	filters := []string{"is_palindrome", "min_length", "max_length", "word_count", "contains_character"}
	stringFilters := make(map[string]string)
	for _, filter := range filters {
		stringFilters[filter] = ""
	}

	//establish DB connection
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("unable to establish connection to database: %v", err)
	}

	//setup state for API
	dbQueries := database.New(db)
	apiConfiguration := apiConfig{
		DB:           dbQueries,
		QueryFilters: stringFilters,
	}

	//server setup
	port := os.Getenv("PORT")
	mux := http.NewServeMux()
	mux.HandleFunc("GET /strings/{string_value}", apiConfiguration.GetText)
	mux.HandleFunc("GET /strings", apiConfiguration.GetFilteredTexts)
	mux.HandleFunc("GET /strings/filter-by-natural-language", apiConfiguration.GetTexByNaturalLang)
	mux.HandleFunc("POST /strings", apiConfiguration.CreateText)
	mux.HandleFunc("DELETE /strings/{string_value}", apiConfiguration.DeleteText)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("server running on port: %v\n", port)
	log.Fatal(server.ListenAndServe())
}

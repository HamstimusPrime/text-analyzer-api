package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/HamstimusPrime/text-analyzer-api/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	DB *database.Queries
}

type RequestBody struct {
	Value string `json:"value"`
}

type ResponseBody struct {
	Body             string    `json:"body"`
	Email            string    `json:"email"`
	UserID           uuid.UUID `json:"user_id"`
	Password         string    `json:"password"`
	ExpiresInSeconds int       `json:"expires_in_seconds"`
}

func (cfg *apiConfig) CreateText(w http.ResponseWriter, r *http.Request) {
	// Parse Request Body
	reqBody, err := parseReqBody(r, RequestBody{})
	if err != nil {
		log.Fatalf("unable to parse request body, err: %v\n", err)

		errMsg := "Internal server error!"
		respondWithError(w, errMsg, 500)
		return
	}

	errorCode, errMsg, err := validateString(reqBody, cfg)
	if err != nil {
		respondWithError(w, errMsg, errorCode)
		return
	}

	createTextParams := database.CreateTextParams{
		Value:        reqBody.Value,
		Length:       charCount(reqBody.Value),
		IsPalindrome: isPalindrome(reqBody.Value),
		WordCount:    wordCount(reqBody.Value),
		Sha256Hash:   generateHash(reqBody.Value),
	}

	stringID, err := cfg.DB.CreateText(context.Background(), createTextParams)
	if err != nil {
		fmt.Printf("error: %v", err)
		errMsg := "unable to save Text to DB"
		respondWithError(w, errMsg, 500)
		return
	}

	uniqueChars := countUniqueChars(reqBody.Value)
	for character, charCount := range uniqueChars {
		createCharCountParams := database.CreateCharCountParams{
			StringID:  stringID,
			Character: string(character),
			CharCount: charCount,
		}

		err := cfg.DB.CreateCharCount(context.Background(), createCharCountParams)
		if err != nil {
			cfg.DB.DeleteText(context.Background(), stringID)
			return
		}
	}

}

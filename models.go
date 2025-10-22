package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/HamstimusPrime/text-analyzer-api/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	DB *database.Queries
}

type RequestBody struct {
	Value string `json:"value"`
}

type SuccessResponseFilterBody struct {
	Body             string    `json:"body"`
	Email            string    `json:"email"`
	UserID           uuid.UUID `json:"user_id"`
	Password         string    `json:"password"`
	ExpiresInSeconds int       `json:"expires_in_seconds"`
}

type SuccessResponseBody struct {
	ID         uuid.UUID `json:"id"`
	Value      string    `json:"value"`
	Properties struct {
		Length                int32          `json:"length"`
		IsPalindrome          string         `json:"is_palindrome"`
		UniqueCharacters      string         `json:"unique_characters"`
		WordCount             string         `json:"word_count"`
		Sha256Hash            string         `json:"sha256_hash"`
		CharacterFrequencyMap map[string]int `json:"character_frequency_map"`
	} `json:"properties"`
	CreatedAt time.Time `json:"created_at"`
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
	//setup inputs for Text-string rows
	createTextParams := database.CreateTextParams{
		Value:        reqBody.Value,
		Length:       charCount(reqBody.Value),
		IsPalindrome: isPalindrome(reqBody.Value),
		WordCount:    wordCount(reqBody.Value),
		Sha256Hash:   generateHash(reqBody.Value),
	}
	//on creation of new Text-string, return ID of created string
	stringID, err := cfg.DB.CreateText(context.Background(), createTextParams)
	if err != nil {
		fmt.Printf("error: %v", err)
		errMsg := "unable to save Text to DB"
		respondWithError(w, errMsg, http.StatusInternalServerError)
		return
	}

	//setup inputs for character count rows
	uniqueChars := countUniqueChars(reqBody.Value)
	for character, charCount := range uniqueChars {
		createCharCountParams := database.CreateCharCountParams{
			StringID:        stringID, //use ID of created string as string ID for each character
			Character:       string(character),
			UniqueCharCount: charCount,
		}
		//create entry into char_count table
		err := cfg.DB.CreateCharCount(context.Background(), createCharCountParams)
		if err != nil {
			cfg.DB.DeleteText(context.Background(), stringID)
			return
		}
	}

	//get character counts for the created text
	charCounts, err := cfg.DB.GetCharacterCountsByID(context.Background(), stringID)
	if err != nil {
		fmt.Printf("error: %v", err)
		errMsg := "unable to get character counts from DB"
		respondWithError(w, errMsg, http.StatusInternalServerError)
		return
	}

	//get text information by ID
	textInfo, err := cfg.DB.GetTextByID(context.Background(), stringID)
	if err != nil {
		fmt.Printf("error: %v", err)
		errMsg := "unable to get text info from DB"
		respondWithError(w, errMsg, http.StatusInternalServerError)
		return
	}

	//create character frequency map
	characterFrequencyMap := make(map[string]int)
	for _, charCount := range charCounts {
		characterFrequencyMap[charCount.Character] = int(charCount.UniqueCharCount)
	}

	//create response body with the parsed data
	responseBody := SuccessResponseBody{
		ID:    textInfo.ID,
		Value: textInfo.Value,
		Properties: struct {
			Length                int32          `json:"length"`
			IsPalindrome          string         `json:"is_palindrome"`
			UniqueCharacters      string         `json:"unique_characters"`
			WordCount             string         `json:"word_count"`
			Sha256Hash            string         `json:"sha256_hash"`
			CharacterFrequencyMap map[string]int `json:"character_frequency_map"`
		}{
			Length:                textInfo.Length,
			IsPalindrome:          fmt.Sprintf("%t", textInfo.IsPalindrome),
			UniqueCharacters:      fmt.Sprintf("%d", len(charCounts)),
			WordCount:             fmt.Sprintf("%d", textInfo.WordCount),
			Sha256Hash:            textInfo.Sha256Hash,
			CharacterFrequencyMap: characterFrequencyMap,
		},
		CreatedAt: textInfo.CreatedAt,
	}

	//return JSON response
	fmt.Println("text created!!")
	respondWithJSON(w, responseBody, 200)

}

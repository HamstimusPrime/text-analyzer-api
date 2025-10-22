package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/HamstimusPrime/text-analyzer-api/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	DB           *database.Queries
	QueryFilters map[string]string
}

type RequestBody struct {
	Value string `json:"value"`
}

type FilteredTextsResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Value      string `json:"value"`
		Properties struct {
		} `json:"properties"`
		CreatedAt time.Time `json:"created_at"`
	} `json:"data"`
	Count          int `json:"count"`
	FiltersApplied struct {
		IsPalindrome      bool   `json:"is_palindrome"`
		MinLength         int    `json:"min_length"`
		MaxLength         int    `json:"max_length"`
		WordCount         int    `json:"word_count"`
		ContainsCharacter string `json:"contains_character"`
	} `json:"filters_applied"`
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
			cfg.DB.DeleteTextWithID(context.Background(), stringID)
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

func (cfg *apiConfig) GetText(w http.ResponseWriter, r *http.Request) {
	// Extract the string value from the URL path
	stringValue := r.PathValue("string_value")
	if stringValue == "" {
		errMsg := "String does not exist in the system"
		respondWithError(w, errMsg, http.StatusBadRequest)
		return
	}
	// Get text information by value from database
	textInfo, err := cfg.DB.GetText(context.Background(), stringValue)
	if err != nil {
		if err == sql.ErrNoRows {
			errMsg := "string not found"
			respondWithError(w, errMsg, http.StatusNotFound)
			return
		}
		fmt.Printf("error: %v", err)
		errMsg := "unable to get text info from DB"
		respondWithError(w, errMsg, http.StatusInternalServerError)
		return
	}

	// Get character counts for the text
	charCounts, err := cfg.DB.GetCharacterCountsByID(context.Background(), textInfo.ID)
	if err != nil {
		fmt.Printf("error: %v", err)
		errMsg := "unable to get character counts from DB"
		respondWithError(w, errMsg, http.StatusInternalServerError)
		return
	}

	// Create character frequency map
	characterFrequencyMap := make(map[string]int)
	for _, charCount := range charCounts {
		characterFrequencyMap[charCount.Character] = int(charCount.UniqueCharCount)
	}

	// Create response body with the parsed data
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

	// Return JSON response
	respondWithJSON(w, responseBody, http.StatusOK)
}

func (cfg *apiConfig) GetFilteredTexts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	clientQueryFilters := r.URL.Query()

	//check if keys in  client-query correspond to Querykeys
	for query := range clientQueryFilters {
		if _, ok := cfg.QueryFilters[query]; !ok {
			errMsg := "Invalid query parameter values or types"
			respondWithError(w, errMsg, http.StatusBadRequest)
			return
		}

	}

	// Create params struct for the database query
	filterParams := database.GetFilteredTextsParams{
		// Set default values that won't filter anything
		IsPalindrome:      false, // Will need logic to handle optional boolean
		MinLength:         0,
		MaxLength:         999999, // Large default
		WordCount:         0,      // Will need logic to handle optional int
		ContainsCharacter: sql.NullString{Valid: false},
	}

	// Flags to track which optional parameters were actually provided
	// var isPalindromeProvided, wordCountProvided bool

	// Parse and validate each query parameter
	for key, values := range clientQueryFilters {
		if len(values) == 0 {
			continue
		}
		value := values[0] // Take first value if multiple provided

		switch key {
		case "is_palindrome":
			// Parse boolean and set flag to indicate it was provided
			palindromeVal, err := strconv.ParseBool(value)
			if err != nil {
				errMsg := "Invalid is_palindrome parameter: must be 'true' or 'false'"
				respondWithError(w, errMsg, http.StatusBadRequest)
				return
			}
			filterParams.IsPalindrome = palindromeVal

		case "min_length":
			// Parse int32 and validate > 0
			minLength, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				errMsg := "Invalid min_length parameter: must be a valid integer"
				respondWithError(w, errMsg, http.StatusBadRequest)
				return
			}
			if minLength < 0 {
				errMsg := "Invalid min_length parameter: must be greater than or equal to 0"
				respondWithError(w, errMsg, http.StatusBadRequest)
				return
			}
			filterParams.MinLength = int32(minLength)

		case "max_length":
			// Parse int32 and validate > 0
			maxLength, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				errMsg := "Invalid max_length parameter: must be a valid integer"
				respondWithError(w, errMsg, http.StatusBadRequest)
				return
			}
			if maxLength <= 0 {
				errMsg := "Invalid max_length parameter: must be greater than 0"
				respondWithError(w, errMsg, http.StatusBadRequest)
				return
			}
			filterParams.MaxLength = int32(maxLength)

		case "word_count":
			// Parse int32 and validate >= 0
			wordCount, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				errMsg := "Invalid word_count parameter: must be a valid integer"
				respondWithError(w, errMsg, http.StatusBadRequest)
				return
			}
			if wordCount < 0 {
				errMsg := "Invalid word_count parameter: must be greater than or equal to 0"
				respondWithError(w, errMsg, http.StatusBadRequest)
				return
			}
			filterParams.WordCount = int32(wordCount)

		case "contains_character":
			// Validate string and set NullString
			if strings.TrimSpace(value) == "" {
				errMsg := "Invalid contains_character parameter: cannot be empty or whitespace only"
				respondWithError(w, errMsg, http.StatusBadRequest)
				return
			}
			filterParams.ContainsCharacter = sql.NullString{
				String: value,
				Valid:  true,
			}
		}
	}

	// Call the database function
	texts, err := cfg.DB.GetFilteredTexts(context.Background(), filterParams)
	if err != nil {
		fmt.Printf("error getting filtered texts: %v", err)
		errMsg := "Unable to retrieve filtered texts from database"
		respondWithError(w, errMsg, http.StatusInternalServerError)
		return
	}

	// Parse texts into FilteredTextsResponse struct
	response := FilteredTextsResponse{
		Data: make([]struct {
			ID         string    `json:"id"`
			Value      string    `json:"value"`
			Properties struct{}  `json:"properties"`
			CreatedAt  time.Time `json:"created_at"`
		}, len(texts)),
		Count: len(texts),
		FiltersApplied: struct {
			IsPalindrome      bool   `json:"is_palindrome"`
			MinLength         int    `json:"min_length"`
			MaxLength         int    `json:"max_length"`
			WordCount         int    `json:"word_count"`
			ContainsCharacter string `json:"contains_character"`
		}{
			IsPalindrome:      filterParams.IsPalindrome,
			MinLength:         int(filterParams.MinLength),
			MaxLength:         int(filterParams.MaxLength),
			WordCount:         int(filterParams.WordCount),
			ContainsCharacter: filterParams.ContainsCharacter.String,
		},
	}

	for i, text := range texts {
		response.Data[i].ID = text.ID.String()
		response.Data[i].Value = text.Value
		response.Data[i].CreatedAt = text.CreatedAt
	}

	respondWithJSON(w, response, http.StatusOK)
}

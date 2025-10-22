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
			Length                int32          `json:"length"`
			IsPalindrome          string         `json:"is_palindrome"`
			WordCount             string         `json:"word_count"`
			Sha256Hash            string         `json:"sha256_hash"`
			CharacterFrequencyMap map[string]int `json:"character_frequency_map"`
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

// NLPFilters represents the parsed natural language query
type NLPFilters struct {
	IsPalindrome      *bool   `json:"is_palindrome,omitempty"`
	MinLength         *int    `json:"min_length,omitempty"`
	MaxLength         *int    `json:"max_length,omitempty"`
	WordCount         *int    `json:"word_count,omitempty"`
	ContainsCharacter *string `json:"contains_character,omitempty"`
	ContainsText      *string `json:"contains_text,omitempty"`
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
	uniqueChars := getUniqueChars(reqBody.Value)
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
			UniqueCharacters:      fmt.Sprintf("%d", len(characterFrequencyMap)),
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
			UniqueCharacters:      fmt.Sprintf("%d", len(characterFrequencyMap)),
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
			ID         string `json:"id"`
			Value      string `json:"value"`
			Properties struct {
				Length                int32          `json:"length"`
				IsPalindrome          string         `json:"is_palindrome"`
				WordCount             string         `json:"word_count"`
				Sha256Hash            string         `json:"sha256_hash"`
				CharacterFrequencyMap map[string]int `json:"character_frequency_map"`
			} `json:"properties"`
			CreatedAt time.Time `json:"created_at"`
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
		// Get character counts for each text to build frequency map
		charCounts, err := cfg.DB.GetCharacterCountsByID(context.Background(), text.ID)
		if err != nil {
			fmt.Printf("error getting character counts for text %s: %v", text.ID.String(), err)
			// Continue with empty character frequency map rather than failing completely
			charCounts = []database.GetCharacterCountsByIDRow{}
		}

		// Create character frequency map
		characterFrequencyMap := make(map[string]int)
		for _, charCount := range charCounts {
			characterFrequencyMap[charCount.Character] = int(charCount.UniqueCharCount)
		}

		response.Data[i].ID = text.ID.String()
		response.Data[i].Value = text.Value
		response.Data[i].CreatedAt = text.CreatedAt
		response.Data[i].Properties.Length = text.Length
		response.Data[i].Properties.IsPalindrome = fmt.Sprintf("%t", text.IsPalindrome)
		response.Data[i].Properties.WordCount = fmt.Sprintf("%d", text.WordCount)
		response.Data[i].Properties.Sha256Hash = text.Sha256Hash
		response.Data[i].Properties.CharacterFrequencyMap = characterFrequencyMap
	}

	respondWithJSON(w, response, http.StatusOK)
}

func (cfg *apiConfig) DeleteText(w http.ResponseWriter, r *http.Request) {
	stringValue := r.PathValue("string_value")
	if stringValue == "" {
		errMsg := "String does not exist in the system"
		respondWithError(w, errMsg, http.StatusBadRequest)
		return
	}

	// Get text information by value to get the ID
	textInfo, err := cfg.DB.GetText(context.Background(), stringValue)
	if err != nil {
		if err == sql.ErrNoRows {
			errMsg := "String does not exist in the system"
			respondWithError(w, errMsg, http.StatusNotFound)
			return
		}
		fmt.Printf("error: %v", err)
		errMsg := "unable to get text info from DB"
		respondWithError(w, errMsg, http.StatusInternalServerError)
		return
	}

	// Delete the text by ID
	err = cfg.DB.DeleteTextWithID(context.Background(), textInfo.ID)
	if err != nil {
		fmt.Printf("error deleting text: %v", err)
		errMsg := "unable to delete text from DB"
		respondWithError(w, errMsg, http.StatusInternalServerError)
		return
	}

	// Return success response with 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// func (cfg *apiConfig) GetTexByNaturalLang(w http.ResponseWriter, r *http.Request) {
// 	// Get the natural language query from request body
// 	var reqBody struct {
// 		Query string `json:"query"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
// 		respondWithError(w, "Invalid JSON", http.StatusBadRequest)
// 		return
// 	}

// 	// Parse natural language query into database filters
// 	filters, err := parseNaturalLanguageQuery(reqBody.Query)
// 	if err != nil {
// 		respondWithError(w, fmt.Sprintf("Could not understand query: %s", err.Error()), http.StatusBadRequest)
// 		return
// 	}

// 	// Build and execute database query based on parsed filters
// 	texts, err := cfg.executeFilteredQuery(r.Context(), filters)
// 	if err != nil {
// 		respondWithError(w, "Database query failed", http.StatusInternalServerError)
// 		return
// 	}

// 	// Format and return response
// 	response := FilteredTextsResponse{
// 		Data:           texts,
// 		Count:          len(texts),
// 		FiltersApplied: convertNLPFiltersToResponse(filters),
// 	}

// 	respondWithJSON(w, response, http.StatusOK)
// }

// parseNaturalLanguageQuery converts natural language to database filters
// func parseNaturalLanguageQuery(query string) (NLPFilters, error) {
// 	query = strings.ToLower(strings.TrimSpace(query))
// 	filters := NLPFilters{}

// 	// Pattern matching for different query types

// 	// Palindrome detection
// 	if strings.Contains(query, "palindrome") {
// 		isPalindrome := true
// 		if strings.Contains(query, "not") || strings.Contains(query, "non") {
// 			isPalindrome = false
// 		}
// 		filters.IsPalindrome = &isPalindrome
// 	}

// 	// Length filters
// 	lengthPattern := regexp.MustCompile(`(?:length|characters?|chars?)\s*(?:(?:is\s*)?(?:greater\s+than|more\s+than|over|>)\s*(\d+)|(?:is\s*)?(?:less\s+than|fewer\s+than|under|<)\s*(\d+)|(?:is\s*)?(?:exactly\s*)?(\d+))`)
// 	matches := lengthPattern.FindAllStringSubmatch(query, -1)
// 	for _, match := range matches {
// 		if match[1] != "" { // greater than
// 			if val, err := strconv.Atoi(match[1]); err == nil {
// 				filters.MinLength = &val
// 			}
// 		} else if match[2] != "" { // less than
// 			if val, err := strconv.Atoi(match[2]); err == nil {
// 				filters.MaxLength = &val
// 			}
// 		} else if match[3] != "" { // exactly
// 			if val, err := strconv.Atoi(match[3]); err == nil {
// 				filters.MinLength = &val
// 				filters.MaxLength = &val
// 			}
// 		}
// 	}

// 	// Word count filters
// 	wordPattern := regexp.MustCompile(`(?:words?|word count)\s*(?:(?:is\s*)?(?:greater\s+than|more\s+than|over|>)\s*(\d+)|(?:is\s*)?(?:less\s+than|fewer\s+than|under|<)\s*(\d+)|(?:is\s*)?(?:exactly\s*)?(\d+))`)
// 	wordMatches := wordPattern.FindAllStringSubmatch(query, -1)
// 	for _, match := range wordMatches {
// 		if match[1] != "" || match[2] != "" || match[3] != "" {
// 			// For simplicity, use exact word count matching
// 			if val, err := strconv.Atoi(match[3]); err == nil {
// 				filters.WordCount = &val
// 			}
// 		}
// 	}

// 	// Contains character
// 	charPattern := regexp.MustCompile(`contains?\s+(?:the\s+)?(?:character|char|letter)\s+['"]?([a-zA-Z])['"]?`)
// 	if charMatch := charPattern.FindStringSubmatch(query); len(charMatch) > 1 {
// 		filters.ContainsCharacter = &charMatch[1]
// 	}

// 	// Contains text/substring
// 	textPattern := regexp.MustCompile(`contains?\s+['"]([^'"]+)['"]`)
// 	if textMatch := textPattern.FindStringSubmatch(query); len(textMatch) > 1 {
// 		filters.ContainsText = &textMatch[1]
// 	}

// 	// If no patterns matched, return an error
// 	if filters.IsPalindrome == nil && filters.MinLength == nil &&
// 		filters.MaxLength == nil && filters.WordCount == nil &&
// 		filters.ContainsCharacter == nil && filters.ContainsText == nil {
// 		return filters, fmt.Errorf("could not parse query: '%s'. Try queries like 'palindromes', 'length > 10', 'contains character a', etc.", query)
// 	}

// 	return filters, nil
// }

// // the convertNLPFiltersToResponse function converts NLPFilters to the expected response format
// func convertNLPFiltersToResponse(filters NLPFilters) struct {
// 	IsPalindrome      bool   `json:"is_palindrome"`
// 	MinLength         int    `json:"min_length"`
// 	MaxLength         int    `json:"max_length"`
// 	WordCount         int    `json:"word_count"`
// 	ContainsCharacter string `json:"contains_character"`
// } {
// 	response := struct {
// 		IsPalindrome      bool   `json:"is_palindrome"`
// 		MinLength         int    `json:"min_length"`
// 		MaxLength         int    `json:"max_length"`
// 		WordCount         int    `json:"word_count"`
// 		ContainsCharacter string `json:"contains_character"`
// 	}{}

// 	if filters.IsPalindrome != nil {
// 		response.IsPalindrome = *filters.IsPalindrome
// 	}
// 	if filters.MinLength != nil {
// 		response.MinLength = *filters.MinLength
// 	}
// 	if filters.MaxLength != nil {
// 		response.MaxLength = *filters.MaxLength
// 	}
// 	if filters.WordCount != nil {
// 		response.WordCount = *filters.WordCount
// 	}
// 	if filters.ContainsCharacter != nil {
// 		response.ContainsCharacter = *filters.ContainsCharacter
// 	}

// 	return response
// }

// // executeFilteredQuery runs the database query based on NLP filters
// func (cfg *apiConfig) executeFilteredQuery(ctx context.Context, filters NLPFilters) ([]struct {
// 	ID         string `json:"id"`
// 	Value      string `json:"value"`
// 	Properties struct {
// 		Length                int32          `json:"length"`
// 		IsPalindrome          string         `json:"is_palindrome"`
// 		UniqueCharacters      string         `json:"unique_characters"`
// 		WordCount             string         `json:"word_count"`
// 		Sha256Hash            string         `json:"sha256_hash"`
// 		CharacterFrequencyMap map[string]int `json:"character_frequency_map"`
// 	} `json:"properties"`
// 	CreatedAt time.Time `json:"created_at"`
// }, error) {
// 	// use the existing GetFilteredTexts method where possible
// 	// and fall back to filtering all texts for more complex queries
// 	// Convert NLP filters to the existing filter format where possible
// 	filterParams := database.GetFilteredTextsParams{
// 		IsPalindrome:      false, // default values
// 		MinLength:         0,
// 		MaxLength:         999999,
// 		WordCount:         0,
// 		ContainsCharacter: sql.NullString{Valid: false},
// 	}

// 	useExistingMethod := true

// 	// Check if we can use the existing GetFilteredTexts method
// 	if filters.IsPalindrome != nil {
// 		filterParams.IsPalindrome = *filters.IsPalindrome
// 	}

// 	if filters.MinLength != nil {
// 		filterParams.MinLength = int32(*filters.MinLength)
// 	} else {
// 		filterParams.MinLength = 0
// 	}

// 	if filters.MaxLength != nil {
// 		filterParams.MaxLength = int32(*filters.MaxLength)
// 	} else {
// 		filterParams.MaxLength = 999999
// 	}

// 	if filters.WordCount != nil {
// 		filterParams.WordCount = int32(*filters.WordCount)
// 	} else {
// 		filterParams.WordCount = 0
// 		useExistingMethod = false // WordCount filtering not well supported in existing method
// 	}

// 	if filters.ContainsCharacter != nil {
// 		filterParams.ContainsCharacter = sql.NullString{
// 			String: *filters.ContainsCharacter,
// 			Valid:  true,
// 		}
// 	}

// 	// If ContainsText is specified, we can't use the existing method efficiently
// 	if filters.ContainsText != nil {
// 		useExistingMethod = false
// 	}

// 	if useExistingMethod && filters.WordCount == nil && filters.ContainsText == nil {
// 		// Use the existing method for simple cases
// 		dbTexts, err := cfg.DB.GetFilteredTexts(ctx, filterParams)
// 		if err != nil {
// 			return nil, err
// 		}

// 		var results []struct {
// 			ID         string `json:"id"`
// 			Value      string `json:"value"`
// 			Properties struct {
// 				Length                int32          `json:"length"`
// 				IsPalindrome          string         `json:"is_palindrome"`
// 				UniqueCharacters      string         `json:"unique_characters"`
// 				WordCount             string         `json:"word_count"`
// 				Sha256Hash            string         `json:"sha256_hash"`
// 				CharacterFrequencyMap map[string]int `json:"character_frequency_map"`
// 			} `json:"properties"`
// 			CreatedAt time.Time `json:"created_at"`
// 		}

// 		for _, text := range dbTexts {
// 			// Get character counts for each text to build frequency map
// 			charCounts, err := cfg.DB.GetCharacterCountsByID(ctx, text.ID)
// 			if err != nil {
// 				fmt.Printf("error getting character counts for text %s: %v", text.ID.String(), err)
// 				// Continue with empty character frequency map rather than failing completely
// 				charCounts = []database.GetCharacterCountsByIDRow{}
// 			}

// 			// Create character frequency map
// 			characterFrequencyMap := make(map[string]int)
// 			for _, charCount := range charCounts {
// 				characterFrequencyMap[charCount.Character] = int(charCount.UniqueCharCount)
// 			}

// 			results = append(results, struct {
// 				ID         string `json:"id"`
// 				Value      string `json:"value"`
// 				Properties struct {
// 					Length                int32          `json:"length"`
// 					IsPalindrome          string         `json:"is_palindrome"`
// 					UniqueCharacters      string         `json:"unique_characters"`
// 					WordCount             string         `json:"word_count"`
// 					Sha256Hash            string         `json:"sha256_hash"`
// 					CharacterFrequencyMap map[string]int `json:"character_frequency_map"`
// 				} `json:"properties"`
// 				CreatedAt time.Time `json:"created_at"`
// 			}{
// 				ID:        text.ID.String(),
// 				Value:     text.Value,
// 				CreatedAt: text.CreatedAt,
// 				Properties: struct {
// 					Length                int32          `json:"length"`
// 					IsPalindrome          string         `json:"is_palindrome"`
// 					UniqueCharacters      string         `json:"unique_characters"`
// 					WordCount             string         `json:"word_count"`
// 					Sha256Hash            string         `json:"sha256_hash"`
// 					CharacterFrequencyMap map[string]int `json:"character_frequency_map"`
// 				}{
// 					Length:                text.Length,
// 					IsPalindrome:          fmt.Sprintf("%t", text.IsPalindrome),
// 					UniqueCharacters:      fmt.Sprintf("%d", len(charCounts)),
// 					WordCount:             fmt.Sprintf("%d", text.WordCount),
// 					Sha256Hash:            text.Sha256Hash,
// 					CharacterFrequencyMap: characterFrequencyMap,
// 				},
// 			})
// 		}

// 		return results, nil
// 	}

// 	// For complex queries with text contains or word count,
// 	// we would need to add custom SQL queries to the sqlc schema.
// 	// For now, return a message suggesting this enhancement
// 	return nil, fmt.Errorf("complex filtering with text search or exact word count requires additional SQL queries. Consider adding custom queries to your sqlc schema")
// }

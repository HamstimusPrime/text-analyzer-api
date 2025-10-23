package main

import (
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

// NaturalLanguageResponse represents the response format for natural language queries
type NaturalLanguageResponse struct {
	Data []string `json:"data"`
	Count int `json:"count"`
	InterpretedQuery struct {
		Original string `json:"original"`
		ParsedFilters map[string]interface{} `json:"parsed_filters"`
	} `json:"interpreted_query"`
}


package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"unicode"
)

func isPalindrome(text string) bool {
	isPalindrome := false
	text = strings.ToLower(text)

	for index := range text {
		endIndex := (len(text) - 1) - index
		if index >= endIndex {
			return isPalindrome
		}
		isPalindrome = false
		if text[index] == text[endIndex] {
			isPalindrome = true
		}
	}
	return isPalindrome
}

func parseReqBody(req *http.Request, format RequestBody) (RequestBody, error) {
	if err := json.NewDecoder(req.Body).Decode(&format); err != nil {
		return RequestBody{}, err
	}
	return format, nil
}

func respondWithError(w http.ResponseWriter, errMsg string, HTTPstatus int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(HTTPstatus)
	w.Write([]byte(errMsg))
}

func respondWithJSON(w http.ResponseWriter, resTemplate interface{}, HTTPstatus int) {
	resJSON, err := json.Marshal(resTemplate)
	if err != nil {
		log.Fatal("unable to parse response JSON")
	}
	w.Header().Set("Content-Type", "json/plain; charset=utf-8")
	w.WriteHeader(HTTPstatus)
	w.Write([]byte(resJSON))
}

func validateString(reqBody RequestBody, cfg *apiConfig) (int, string, error) {
	if reqBody.Value == "" || strings.TrimSpace(reqBody.Value) == "" {
		errMsg := fmt.Sprintf(`Invalid request body or missing "value" field`)
		return 400, errMsg, errors.New("no strings passed in value field")
	}

	_, err := strconv.Atoi(reqBody.Value)
	if err == nil {
		errMsg := fmt.Sprintf(`Unprocessable Entity`)
		return 422, errMsg, errors.New("invalid string format")
	}

	//check if String in DB
	text, err := cfg.DB.GetText(context.Background(), reqBody.Value)
	if err != nil {
		// If it's a "not found" error(string doesn't exist) Only return an error if it's a real database error
		if !errors.Is(err, sql.ErrNoRows) {
			errMsg := fmt.Sprintf("Failed to validate text\n")
			fmt.Printf("error: %v", err)
			return 500, errMsg, err
		}
	}

	if text != "" {
		errMsg := fmt.Sprintf("String already exists in the system")
		return 409, errMsg, errors.New("String in DB")
	}
	return 0, "", nil
}

func generateHash(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	return string(h.Sum(nil))
}

func charCount(str string) int32 {
	count := 0
	for _, r := range str {
		if !unicode.IsSpace(r) {
			count++
		}
	}
	return int32(count)
}

func wordCount(str string) int32 {
	words := strings.Fields(str)
	return int32(len(words))
}

// unique character count
func countUniqueChars(s string) map[rune]int32 {
	counts := make(map[rune]int32)

	for _, char := range s {
		counts[char]++
	}

	return counts
}

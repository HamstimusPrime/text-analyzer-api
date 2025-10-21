package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/HamstimusPrime/text-analyzer-api/models"
)

func IsPalindrome(text string) bool {
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

func ParseReqBody(req *http.Request, format models.RequestBody) (models.RequestBody, error) {
	if err := json.NewDecoder(req.Body).Decode(&format); err != nil {
		return models.RequestBody{}, err
	}
	return format, nil
}

func TextExistsInDB() bool {
	//takes the queries from the database and checks if it matches(don't forget to remove white space at ends)
	ret
}

func RespondWithError(w http.ResponseWriter, errMsg string, HTTPstatus int) {
	w.WriteHeader(HTTPstatus)
	errJSON, err := json.Marshal(errorMsg{Error: errMsg})
	if err != nil {
		log.Fatal("unable to parse error JSON")
	}
	w.Write([]byte(errJSON))
}

func RespondWithJSON(w http.ResponseWriter, resTemplate interface{}, HTTPstatus int) {
	resJSON, err := json.Marshal(resTemplate)
	if err != nil {
		log.Fatal("unable to parse response JSON")
	}
	w.Header().Set("Content-Type", "json/plain; charset=utf-8")
	w.WriteHeader(HTTPstatus)
	w.Write([]byte(resJSON))
}

func ValidateString(reqBody models.RequestBody) string {
	text, ok := reqBody.Value
	if !ok {
		return "Value must be a string"
	}
	if strings.TrimSpace(text) == "" {
		return "Value cannot be empty"
	}
	return ""
}

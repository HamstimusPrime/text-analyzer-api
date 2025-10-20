package utils

func IsPalindrome(text string) bool {
	isPalindrome := false

	for index, _ := range text {
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

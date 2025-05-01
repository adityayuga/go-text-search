package gotextsearch

import "strings"

// Tokenize splits the input text into tokens (words).
func Tokenize(text string) []string {
	var tokens []string
	for _, word := range strings.Fields(text) {
		tokens = append(tokens, word)
	}
	return tokens
}

// RemoveSymbols removes all non-alphanumeric characters from the input text.
func RemoveSymbols(text string) string {
	// Remove all non-alphanumeric characters
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, text)
}

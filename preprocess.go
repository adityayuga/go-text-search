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

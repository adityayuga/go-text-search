package gotextsearch

// autoCorrectQueryTerm will return the best match for a term based on Levenshtein distance
func (idx *invertedIndex) autoCorrectQueryTerm(term string) string {
	best := term
	maxDist := 2
	bestDist := maxDist + 1
	for known := range idx.vocab {
		d := levenshtein(term, known)
		if d < bestDist {
			best = known
			bestDist = d
			if d == 1 {
				break
			}
		}
	}
	return best
}

// levenshtein distance algorithm
func levenshtein(a, b string) int {
	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}

	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			matrix[i][j] = minFromThree(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}
	return matrix[len(a)][len(b)]
}

// minFromThree returns the minimum of three integers
func minFromThree(a, b, c int) int {
	if a < b && a < c {
		return a
	} else if b < c {
		return b
	}
	return c
}

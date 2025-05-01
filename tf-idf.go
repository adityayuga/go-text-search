package gotextsearch

import (
	"math"
	"sort"
	"strings"
)

type (
	document struct {
		id   int
		text string
	}

	posting struct {
		docID int
		tf    float64
	}

	invertedIndex struct {
		listDocs   map[int]document
		index      map[string][]posting
		docFreq    map[string]int
		docCount   int
		docLengths map[int]float64
		vocab      map[string]bool

		// configuration
		stopWords                       map[string]bool
		preprocessor                    []func(string) string
		tokenizer                       func(string) []string
		autoCorrectQueryTermMaxDistance int
	}

	Config struct {
		StopWords    []string
		Preprocessor []func(string) string
		Tokenizer    func(string) []string
	}

	SearchResult struct {
		Text  string
		Score float64
	}

	SearchIndex interface {
		AddDocument(text string)
		SeedDocuments(texts []string)
		ComputeDocLengths()
		Search(query string, limit int) []SearchResult
	}
)

// New create a new inverted index with the default configuration
func New() SearchIndex {
	return NewWithConfig(Config{})
}

// NewWithConfig create a new inverted index with the given configuration
func NewWithConfig(cfg Config) SearchIndex {
	if cfg.Tokenizer == nil {
		cfg.Tokenizer = Tokenize
	}

	listPreprocessor := []func(string) string{
		strings.ToLower,
	}

	for _, preprocessor := range cfg.Preprocessor {
		listPreprocessor = append(listPreprocessor, preprocessor)
	}

	if cfg.StopWords == nil {
		cfg.StopWords = []string{}
	}

	idx := &invertedIndex{
		listDocs:   make(map[int]document),
		index:      make(map[string][]posting),
		docFreq:    make(map[string]int),
		docCount:   0,
		docLengths: make(map[int]float64),
		vocab:      make(map[string]bool),

		preprocessor:                    cfg.Preprocessor,
		tokenizer:                       cfg.Tokenizer,
		autoCorrectQueryTermMaxDistance: 2,
	}

	idx.stopWords = make(map[string]bool)
	for _, stopWord := range cfg.StopWords {
		idx.stopWords[stopWord] = true
	}

	return idx
}

// AddDocument AddDocument will add a document to the index, automatically computing the term frequency
// need to call ComputeDocLengths after adding documents
func (idx *invertedIndex) AddDocument(text string) {
	// run proprocessor
	for _, preprocessor := range idx.preprocessor {
		text = preprocessor(text)
	}

	// run tokenizer
	tokens := idx.tokenizer(text)

	// remove stopwords
	filteredTokens := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if _, ok := idx.stopWords[token]; !ok {
			filteredTokens = append(filteredTokens, token)
		}
	}

	// dont proceed if no tokens
	if len(filteredTokens) == 0 {
		return
	}

	// each of new docs, will automatically assigned with new incremental ID
	idx.docCount++

	// append docs
	docID := idx.docCount
	doc := document{id: docID, text: text}
	idx.listDocs[docID] = doc

	// process tokens on the doc
	tfCount := map[string]int{}
	for _, token := range filteredTokens {
		tfCount[token]++
		idx.vocab[token] = true
	}

	// update index
	for term, count := range tfCount {
		idx.docFreq[term]++
		tf := float64(count) / float64(len(filteredTokens))
		idx.index[term] = append(idx.index[term], posting{docID: idx.docCount, tf: tf})
	}
}

// SeedDocuments will add documents to the index, automatically compute doc lengths
func (idx *invertedIndex) SeedDocuments(texts []string) {
	for _, text := range texts {
		idx.AddDocument(text)
	}

	idx.ComputeDocLengths()
}

// ComputeDocLengths computes the length of each document in the index
func (idx *invertedIndex) ComputeDocLengths() {
	idx.docLengths = make(map[int]float64)
	for term, postings := range idx.index {
		idf := math.Log(float64(idx.docCount) / (1 + float64(idx.docFreq[term])))
		for _, p := range postings {
			w := p.tf * idf
			idx.docLengths[p.docID] += w * w
		}
	}
	for docID := range idx.docLengths {
		idx.docLengths[docID] = math.Sqrt(idx.docLengths[docID])
	}
}

// Search will search the index for the given query and return the top N results
func (idx *invertedIndex) Search(query string, limit int) []SearchResult {
	// preprocess query and calculate tf each terms
	queryTerms := idx.preprocessQuery(query)
	queryTF := make(map[string]float64)
	for _, term := range queryTerms {
		queryTF[term]++
	}
	for term := range queryTF {
		queryTF[term] /= float64(len(queryTerms))
	}

	// calculate query norm
	var queryNorm float64
	for term, tf := range queryTF {
		idf := math.Log(float64(idx.docCount) / (1 + float64(idx.docFreq[term])))
		queryNorm += (tf * idf) * (tf * idf)
	}
	queryNorm = math.Sqrt(queryNorm)

	// calculate scores
	scores := make(map[int]float64)
	for term, tf := range queryTF {
		idf := math.Log(float64(idx.docCount) / (1 + float64(idx.docFreq[term])))
		for _, p := range idx.index[term] {
			scores[p.docID] += (tf * idf) * (p.tf * idf)
		}
	}

	var results []SearchResult
	for docID, score := range scores {
		docNorm := idx.docLengths[docID]
		if docNorm == 0 || queryNorm == 0 {
			continue
		}
		score /= (docNorm * queryNorm)
		results = append(results, SearchResult{
			Text:  idx.listDocs[docID].text,
			Score: score,
		})
	}

	// sort results by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// limit results
	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}

	return results
}

// preprocessQuery will run the preprocessor and tokenizer on the query
func (idx *invertedIndex) preprocessQuery(query string) []string {
	// run proprocessor
	for _, preprocessor := range idx.preprocessor {
		query = preprocessor(query)
	}

	// run tokenizer
	tokens := idx.tokenizer(query)

	var corrected []string
	for _, t := range tokens {
		if idx.vocab[t] {
			corrected = append(corrected, t)
		} else {
			corrected = append(corrected, idx.autoCorrectQueryTerm(t, idx.autoCorrectQueryTermMaxDistance))
		}
	}
	return corrected
}

// autoCorrectQueryTerm will return the best match for a term based on Levenshtein distance
func (idx *invertedIndex) autoCorrectQueryTerm(term string, maxDist int) string {
	best := term
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

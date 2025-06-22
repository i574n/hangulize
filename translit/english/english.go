/*
Package english implements a transliterator for English.
It uses a dictionary-based approach to look up the phonetic
pronunciation of a word (in ARPAbet) before transcription.
*/
package english

import (
	"bufio"
	_ "embed" // Required for go:embed
	"io"
	"strings"
	"sync"

	"github.com/hangulize/hangulize"
)

//go:embed cmudict.dict
var cmudict string

// T is a hangulize.Translit for English.
var T hangulize.Translit = &english{}

//------------------------------------------------------------------------------

var (
	dict map[string]string
	once sync.Once
)

type english struct{}

func (english) Scheme() string {
	return "english"
}

// loadDictionary parses a pronunciation dictionary.
func loadDictionary(r io.Reader) (map[string]string, error) {
	dict := make(map[string]string)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, ";;;") {
			continue
		}
		// The dictionary format separates the word and pronunciation with 2 spaces.
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) == 2 {
			word := strings.ToLower(parts[0])
			pron := parts[1]
			// Remove stress numbers from phonemes for simplicity (e.g., "AH0" -> "AH").
			pron = strings.NewReplacer("0", "", "1", "", "2", "", " ", "").Replace(pron)
			dict[word] = pron
		}
	}
	return dict, scanner.Err()
}

// Transliterate converts an English word to its phonetic representation.
func (p *english) Transliterate(word string) (string, error) {
	// Lazily load the embedded dictionary once.
	once.Do(func() {
		r := strings.NewReader(cmudict)
		dict, _ = loadDictionary(r)
	})

	words := strings.Fields(word)
	var result []string

	for _, w := range words {
		// Clean and lowercase the word for dictionary lookup.
		cleanWord := strings.ToLower(strings.Trim(w, ".,!?;:\"'()"))

		if phonetic, ok := dict[cleanWord]; ok {
			result = append(result, phonetic)
		} else {
			// If a word is not in the dictionary, pass it through as is.
			result = append(result, w)
		}
	}

	return strings.Join(result, " "), nil
}

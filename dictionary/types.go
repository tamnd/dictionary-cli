package dictionary

import "strings"

// Definition is the record emitted for definitions.
type Definition struct {
	Rank         int    `json:"rank"`
	Word         string `json:"word"`
	PartOfSpeech string `json:"part_of_speech"`
	Definition   string `json:"definition"`
	Example      string `json:"example"`
	Synonyms     string `json:"synonyms"`
}

// Phonetic is the record emitted for phonetics.
type Phonetic struct {
	Text  string `json:"text"`
	Audio string `json:"audio"`
}

// Synonym is the record emitted for synonyms and antonyms.
type Synonym struct {
	Rank int    `json:"rank"`
	Word string `json:"word"`
	URL  string `json:"url"`
}

// Example is the record emitted for example sentences.
type Example struct {
	Rank    int    `json:"rank"`
	Word    string `json:"word"`
	Example string `json:"example"`
}

// ─── wire types ──────────────────────────────────────────────────────────────

type wireEntry struct {
	Word      string        `json:"word"`
	Phonetic  string        `json:"phonetic"`
	Phonetics []wirePhone   `json:"phonetics"`
	Origin    string        `json:"origin"`
	Meanings  []wireMeaning `json:"meanings"`
}

type wirePhone struct {
	Text  string `json:"text"`
	Audio string `json:"audio"`
}

type wireMeaning struct {
	PartOfSpeech string    `json:"partOfSpeech"`
	Definitions  []wireDef `json:"definitions"`
	Synonyms     []string  `json:"synonyms"`
	Antonyms     []string  `json:"antonyms"`
}

type wireDef struct {
	Definition string   `json:"definition"`
	Example    string   `json:"example"`
	Synonyms   []string `json:"synonyms"`
	Antonyms   []string `json:"antonyms"`
}

type wireError struct {
	Title      string `json:"title"`
	Message    string `json:"message"`
	Resolution string `json:"resolution"`
}

// ─── transformation helpers ───────────────────────────────────────────────────

func entriesToDefinitions(entries []wireEntry, limit int) []Definition {
	var out []Definition
	rank := 1
	for _, e := range entries {
		for _, m := range e.Meanings {
			for _, d := range m.Definitions {
				out = append(out, Definition{
					Rank:         rank,
					Word:         e.Word,
					PartOfSpeech: m.PartOfSpeech,
					Definition:   d.Definition,
					Example:      d.Example,
					Synonyms:     strings.Join(d.Synonyms, ";"),
				})
				rank++
				if limit > 0 && len(out) >= limit {
					return out
				}
			}
		}
	}
	return out
}

func collectWords(entries []wireEntry, field string, lang, baseURL string, limit int) []Synonym {
	seen := map[string]bool{}
	var out []Synonym
	rank := 1

	add := func(words []string) {
		for _, w := range words {
			if w == "" {
				continue
			}
			key := strings.ToLower(w)
			if seen[key] {
				continue
			}
			seen[key] = true
			u := baseURL + "/api/v2/entries/" + lang + "/" + w
			out = append(out, Synonym{Rank: rank, Word: w, URL: u})
			rank++
			if limit > 0 && len(out) >= limit {
				return
			}
		}
	}

	for _, e := range entries {
		for _, m := range e.Meanings {
			if field == "synonyms" {
				add(m.Synonyms)
			} else {
				add(m.Antonyms)
			}
			if limit > 0 && len(out) >= limit {
				return out
			}
			for _, d := range m.Definitions {
				if field == "synonyms" {
					add(d.Synonyms)
				} else {
					add(d.Antonyms)
				}
				if limit > 0 && len(out) >= limit {
					return out
				}
			}
		}
	}
	return out
}

func entriesToPhonetics(entries []wireEntry) []Phonetic {
	var out []Phonetic
	for _, e := range entries {
		for _, p := range e.Phonetics {
			if p.Text == "" && p.Audio == "" {
				continue
			}
			out = append(out, Phonetic(p))
		}
	}
	return out
}

func entriesToExamples(entries []wireEntry, limit int) []Example {
	seen := map[string]bool{}
	var out []Example
	rank := 1
	for _, e := range entries {
		for _, m := range e.Meanings {
			for _, d := range m.Definitions {
				ex := strings.TrimSpace(d.Example)
				if ex == "" {
					continue
				}
				key := strings.ToLower(ex)
				if seen[key] {
					continue
				}
				seen[key] = true
				out = append(out, Example{Rank: rank, Word: e.Word, Example: ex})
				rank++
				if limit > 0 && len(out) >= limit {
					return out
				}
			}
		}
	}
	return out
}

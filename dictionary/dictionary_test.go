package dictionary

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// canned JSON for a word with two meanings.
const helloJSON = `[{
  "word": "hello",
  "phonetic": "həˈloʊ",
  "phonetics": [
    {"text": "həˈloʊ", "audio": "https://example.com/hello-au.mp3"},
    {"text": "", "audio": ""},
    {"text": "hɛˈloʊ", "audio": ""}
  ],
  "origin": "Early 19th century.",
  "meanings": [
    {
      "partOfSpeech": "exclamation",
      "definitions": [
        {
          "definition": "Used as a greeting or to begin a phone conversation.",
          "example": "hello there, Katie!",
          "synonyms": ["hi", "hey"],
          "antonyms": ["goodbye"]
        },
        {
          "definition": "Used to express surprise.",
          "example": "hello, what's all this then?",
          "synonyms": [],
          "antonyms": []
        }
      ],
      "synonyms": ["howdy"],
      "antonyms": ["farewell"]
    },
    {
      "partOfSpeech": "noun",
      "definitions": [
        {
          "definition": "An utterance of 'hello'.",
          "example": "she was all smiles and hellos",
          "synonyms": ["greeting"],
          "antonyms": []
        }
      ],
      "synonyms": [],
      "antonyms": []
    }
  ]
}]`

const notFoundJSON = `{"title":"No Definitions Found","message":"Sorry pal, we couldn't find definitions for the word you were looking for.","resolution":"You can try the search again at later time or head to the web instead."}`

func newTestClient(ts *httptest.Server) *Client {
	cfg := DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return NewClient(cfg)
}

func TestGetSendsUserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write([]byte(helloJSON))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	defs, err := c.Define(context.Background(), "en", "hello", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(defs) == 0 {
		t.Error("expected definitions, got none")
	}
}

func TestDefine(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(helloJSON))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	defs, err := c.Define(context.Background(), "en", "hello", 0)
	if err != nil {
		t.Fatal(err)
	}
	// 2 exclamation defs + 1 noun def = 3 total
	if len(defs) != 3 {
		t.Fatalf("want 3 defs, got %d", len(defs))
	}
	if defs[0].Rank != 1 {
		t.Errorf("first rank = %d, want 1", defs[0].Rank)
	}
	if defs[0].Word != "hello" {
		t.Errorf("word = %q, want hello", defs[0].Word)
	}
	if defs[0].PartOfSpeech != "exclamation" {
		t.Errorf("pos = %q, want exclamation", defs[0].PartOfSpeech)
	}
	if defs[2].PartOfSpeech != "noun" {
		t.Errorf("third pos = %q, want noun", defs[2].PartOfSpeech)
	}
	if defs[2].Rank != 3 {
		t.Errorf("third rank = %d, want 3", defs[2].Rank)
	}
}

func TestDefineLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(helloJSON))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	defs, err := c.Define(context.Background(), "en", "hello", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(defs) != 2 {
		t.Fatalf("want 2 defs (limited), got %d", len(defs))
	}
}

func TestDefineNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(notFoundJSON))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Define(context.Background(), "en", "xyzzy404", 0)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestSynonyms(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(helloJSON))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	syns, err := c.Synonyms(context.Background(), "en", "hello", 0)
	if err != nil {
		t.Fatal(err)
	}
	// meaning-level: "howdy"; def-level: "hi", "hey", "greeting"
	if len(syns) < 1 {
		t.Fatal("expected synonyms, got none")
	}
	// No duplicates — "hi" appears only once even if it shows up at multiple levels.
	seen := map[string]bool{}
	for _, s := range syns {
		lower := s.Word
		if seen[lower] {
			t.Errorf("duplicate synonym: %q", s.Word)
		}
		seen[lower] = true
		if s.URL == "" {
			t.Errorf("synonym %q has empty URL", s.Word)
		}
	}
}

func TestAntonyms(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(helloJSON))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	ants, err := c.Antonyms(context.Background(), "en", "hello", 0)
	if err != nil {
		t.Fatal(err)
	}
	// meaning-level: "farewell"; def-level: "goodbye"
	if len(ants) < 1 {
		t.Fatal("expected antonyms, got none")
	}
}

func TestPhonetics(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(helloJSON))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	phones, err := c.Phonetics(context.Background(), "en", "hello")
	if err != nil {
		t.Fatal(err)
	}
	// canned JSON has 3 entries: first has text+audio, second both empty (skipped), third has text only
	if len(phones) != 2 {
		t.Fatalf("want 2 phonetics (skip empty), got %d", len(phones))
	}
	if phones[0].Text != "həˈloʊ" {
		t.Errorf("first text = %q, want həˈloʊ", phones[0].Text)
	}
	if phones[0].Audio == "" {
		t.Error("first audio should not be empty")
	}
}

func TestExamples(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(helloJSON))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	examples, err := c.Examples(context.Background(), "en", "hello", 0)
	if err != nil {
		t.Fatal(err)
	}
	// 3 definitions each with an example = 3 unique examples
	if len(examples) != 3 {
		t.Fatalf("want 3 examples, got %d", len(examples))
	}
	if examples[0].Rank != 1 {
		t.Errorf("first rank = %d, want 1", examples[0].Rank)
	}
	if examples[0].Word != "hello" {
		t.Errorf("word = %q, want hello", examples[0].Word)
	}
}

func TestRetryOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(helloJSON))
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := NewClient(cfg)

	start := time.Now()
	defs, err := c.Define(context.Background(), "en", "hello", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(defs) == 0 {
		t.Error("expected definitions after retry")
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("retries did not back off")
	}
}

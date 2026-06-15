// Package dictionary is the library behind the dict command: the HTTP client,
// request shaping, and the typed data models for the Free Dictionary API.
//
// The API at api.dictionaryapi.dev is completely free and requires no
// authentication. The client sets a real User-Agent, paces requests, and
// retries 429/5xx with backoff.
package dictionary

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Host is the canonical hostname for the Free Dictionary API.
const Host = "api.dictionaryapi.dev"

// DefaultUserAgent identifies the client to the API.
const DefaultUserAgent = "dict/dev (+https://github.com/tamnd/dictionary-cli)"

// ErrNotFound is returned when the API responds with "No Definitions Found".
var ErrNotFound = errors.New("not found")

// Config holds constructor parameters.
type Config struct {
	BaseURL   string        // default "https://api.dictionaryapi.dev"
	UserAgent string        // default DefaultUserAgent
	Rate      time.Duration // default 100ms
	Retries   int           // default 3
	Timeout   time.Duration // default 15s
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://api.dictionaryapi.dev",
		UserAgent: DefaultUserAgent,
		Rate:      100 * time.Millisecond,
		Retries:   3,
		Timeout:   15 * time.Second,
	}
}

// Client talks to the Free Dictionary API.
type Client struct {
	httpClient *http.Client
	userAgent  string
	baseURL    string
	rate       time.Duration
	retries    int
	mu         sync.Mutex
	last       time.Time
}

// NewClient returns a Client with the given config.
func NewClient(cfg Config) *Client {
	base := strings.TrimRight(cfg.BaseURL, "/")
	return &Client{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		userAgent:  cfg.UserAgent,
		baseURL:    base,
		rate:       cfg.Rate,
		retries:    cfg.Retries,
	}
}

// get fetches a URL with pacing and retries.
func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, rawURL)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", rawURL, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, true, err
	}

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusNotFound {
		// Try to decode the error body.
		var we wireError
		if jerr := json.Unmarshal(b, &we); jerr == nil && we.Title != "" {
			return nil, false, ErrNotFound
		}
		return nil, false, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	return b, false, nil
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.rate <= 0 {
		return
	}
	if wait := c.rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

// getJSON fetches and JSON-decodes into v. Returns ErrNotFound on 404.
func (c *Client) getJSON(ctx context.Context, rawURL string, v any) error {
	body, err := c.get(ctx, rawURL)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("decode %s: %w", rawURL, err)
	}
	return nil
}

// entries fetches the raw wire entries for a word in a language.
func (c *Client) entries(ctx context.Context, lang, word string) ([]wireEntry, error) {
	rawURL := c.baseURL + "/api/v2/entries/" + lang + "/" + url.PathEscape(word)
	var entries []wireEntry
	if err := c.getJSON(ctx, rawURL, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// Define returns flattened Definition records for word in lang.
// If limit > 0, at most limit records are returned.
func (c *Client) Define(ctx context.Context, lang, word string, limit int) ([]Definition, error) {
	entries, err := c.entries(ctx, lang, word)
	if err != nil {
		return nil, fmt.Errorf("define %q: %w", word, err)
	}
	return entriesToDefinitions(entries, limit), nil
}

// Synonyms returns deduplicated Synonym records for word in lang.
// If limit > 0, at most limit records are returned.
func (c *Client) Synonyms(ctx context.Context, lang, word string, limit int) ([]Synonym, error) {
	entries, err := c.entries(ctx, lang, word)
	if err != nil {
		return nil, fmt.Errorf("synonyms %q: %w", word, err)
	}
	return collectWords(entries, "synonyms", lang, c.baseURL, limit), nil
}

// Antonyms returns deduplicated Synonym records (antonyms) for word in lang.
// If limit > 0, at most limit records are returned.
func (c *Client) Antonyms(ctx context.Context, lang, word string, limit int) ([]Synonym, error) {
	entries, err := c.entries(ctx, lang, word)
	if err != nil {
		return nil, fmt.Errorf("antonyms %q: %w", word, err)
	}
	return collectWords(entries, "antonyms", lang, c.baseURL, limit), nil
}

// Phonetics returns Phonetic records for word in lang.
func (c *Client) Phonetics(ctx context.Context, lang, word string) ([]Phonetic, error) {
	entries, err := c.entries(ctx, lang, word)
	if err != nil {
		return nil, fmt.Errorf("phonetics %q: %w", word, err)
	}
	return entriesToPhonetics(entries), nil
}

// Examples returns deduplicated Example records for word in lang.
// If limit > 0, at most limit records are returned.
func (c *Client) Examples(ctx context.Context, lang, word string, limit int) ([]Example, error) {
	entries, err := c.entries(ctx, lang, word)
	if err != nil {
		return nil, fmt.Errorf("examples %q: %w", word, err)
	}
	return entriesToExamples(entries, limit), nil
}

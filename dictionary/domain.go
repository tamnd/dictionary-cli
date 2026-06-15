package dictionary

import (
	"context"
	"errors"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

func init() { kit.Register(Domain{}) }

// Domain is the Free Dictionary API driver. It carries no state; the per-run
// client is built by the factory Register hands kit.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against,
// and the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "dictionary",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "dictionary",
			Short:  "A command line for dictionary lookups.",
			Long: `A command line for dictionary lookups.

Look up definitions, synonyms, antonyms, phonetics, and usage examples for
English words using the Free Dictionary API. No API key required.`,
			Site: "https://dictionaryapi.dev",
			Repo: "https://github.com/tamnd/dictionary-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{Name: "search", Group: "lookup", List: true,
		URIType: "definition", Summary: "Look up word definitions",
		Args: []kit.Arg{{Name: "word", Help: "word to look up"}}}, searchWord)

	kit.Handle(app, kit.OpMeta{Name: "synonyms", Group: "lookup", List: true,
		URIType: "synonym", Summary: "Look up synonyms for a word",
		Args: []kit.Arg{{Name: "word", Help: "word to look up"}}}, synonymsWord)

	kit.Handle(app, kit.OpMeta{Name: "antonyms", Group: "lookup", List: true,
		URIType: "antonym", Summary: "Look up antonyms for a word",
		Args: []kit.Arg{{Name: "word", Help: "word to look up"}}}, antonymsWord)

	kit.Handle(app, kit.OpMeta{Name: "phonetics", Group: "lookup", List: true,
		URIType: "phonetic", Summary: "Show phonetic transcriptions for a word",
		Args: []kit.Arg{{Name: "word", Help: "word to look up"}}}, phoneticsWord)

	kit.Handle(app, kit.OpMeta{Name: "examples", Group: "lookup", List: true,
		URIType: "example", Summary: "Show usage examples for a word",
		Args: []kit.Arg{{Name: "word", Help: "word to look up"}}}, examplesWord)
}

// newClient builds the client from the host-resolved config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	dc := DefaultConfig()
	if cfg.UserAgent != "" {
		dc.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		dc.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		dc.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		dc.Timeout = cfg.Timeout
	}
	return NewClient(dc), nil
}

// wordInput is the shared input struct for all word-lookup operations.
type wordInput struct {
	Word   string  `kit:"arg"    help:"word to look up"`
	Lang   string  `kit:"flag"   help:"language code" default:"en"`
	Limit  int     `kit:"flag,inherit" help:"max results" default:"0"`
	Client *Client `kit:"inject"`
}

func searchWord(ctx context.Context, in wordInput, emit func(Definition) error) error {
	if in.Word == "" {
		return errs.Usage("word required")
	}
	defs, err := in.Client.Define(ctx, in.Lang, in.Word, in.Limit)
	if err != nil {
		return mapErr(err)
	}
	for _, d := range defs {
		if err := emit(d); err != nil {
			return err
		}
	}
	return nil
}

func synonymsWord(ctx context.Context, in wordInput, emit func(Synonym) error) error {
	if in.Word == "" {
		return errs.Usage("word required")
	}
	syns, err := in.Client.Synonyms(ctx, in.Lang, in.Word, in.Limit)
	if err != nil {
		return mapErr(err)
	}
	for _, s := range syns {
		if err := emit(s); err != nil {
			return err
		}
	}
	return nil
}

func antonymsWord(ctx context.Context, in wordInput, emit func(Synonym) error) error {
	if in.Word == "" {
		return errs.Usage("word required")
	}
	ants, err := in.Client.Antonyms(ctx, in.Lang, in.Word, in.Limit)
	if err != nil {
		return mapErr(err)
	}
	for _, a := range ants {
		if err := emit(a); err != nil {
			return err
		}
	}
	return nil
}

func phoneticsWord(ctx context.Context, in wordInput, emit func(Phonetic) error) error {
	if in.Word == "" {
		return errs.Usage("word required")
	}
	phones, err := in.Client.Phonetics(ctx, in.Lang, in.Word)
	if err != nil {
		return mapErr(err)
	}
	for _, p := range phones {
		if err := emit(p); err != nil {
			return err
		}
	}
	return nil
}

func examplesWord(ctx context.Context, in wordInput, emit func(Example) error) error {
	if in.Word == "" {
		return errs.Usage("word required")
	}
	examples, err := in.Client.Examples(ctx, in.Lang, in.Word, in.Limit)
	if err != nil {
		return mapErr(err)
	}
	for _, e := range examples {
		if err := emit(e); err != nil {
			return err
		}
	}
	return nil
}

// mapErr translates well-known client errors to kit error kinds.
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrNotFound) {
		return errs.NotFound("word not found")
	}
	return err
}

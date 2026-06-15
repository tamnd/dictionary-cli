package dictionary

import (
	"testing"

	"github.com/tamnd/any-cli/kit"
)

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "dictionary" {
		t.Errorf("Scheme = %q, want dictionary", info.Scheme)
	}
	if info.Identity.Binary != "dictionary" {
		t.Errorf("Identity.Binary = %q, want dictionary", info.Identity.Binary)
	}
	if len(info.Hosts) == 0 {
		t.Error("Hosts must not be empty")
	}
	if info.Hosts[0] != Host {
		t.Errorf("Hosts[0] = %q, want %q", info.Hosts[0], Host)
	}
}

func TestDomainRegistered(t *testing.T) {
	// kit.Open resolves to the registered domain; the init() in domain.go
	// registers "dictionary", so this verifies the registration succeeded.
	_, err := kit.Open()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDomainOps(t *testing.T) {
	// Register installs five operations.
	app := kit.New(Domain{}.Info().Identity)
	Domain{}.Register(app)
	ops := app.Ops()
	want := map[string]bool{
		"search":    true,
		"synonyms":  true,
		"antonyms":  true,
		"phonetics": true,
		"examples":  true,
	}
	for _, op := range ops {
		delete(want, op.Meta().Name)
	}
	for name := range want {
		t.Errorf("missing operation: %q", name)
	}
}

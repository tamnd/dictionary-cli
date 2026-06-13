package cli

import (
	"errors"

	"github.com/tamnd/dictionary-cli/dictionary"
)

func isNotFound(err error) bool {
	return errors.Is(err, dictionary.ErrNotFound)
}

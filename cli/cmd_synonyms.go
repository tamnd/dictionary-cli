package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) synonymsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "synonyms <word>",
		Short: "List synonyms for a word",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			word := args[0]
			n := a.effectiveLimit(0)
			a.progressf("fetching synonyms for %q...", word)
			syns, err := a.client.Synonyms(cmd.Context(), a.lang, word, n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(syns, len(syns))
		},
	}
}

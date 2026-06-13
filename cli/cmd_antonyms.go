package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) antonymsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "antonyms <word>",
		Short: "List antonyms for a word",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			word := args[0]
			n := a.effectiveLimit(0)
			a.progressf("fetching antonyms for %q...", word)
			ants, err := a.client.Antonyms(cmd.Context(), a.lang, word, n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(ants, len(ants))
		},
	}
}

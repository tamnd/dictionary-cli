package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) phoneticCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "phonetic <word>",
		Short: "Show phonetics and audio URLs for a word",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			word := args[0]
			a.progressf("fetching phonetics for %q...", word)
			phones, err := a.client.Phonetics(cmd.Context(), a.lang, word)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(phones, len(phones))
		},
	}
}

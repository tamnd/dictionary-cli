package cli

import (
	"github.com/spf13/cobra"
)

func (a *App) examplesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "examples <word>",
		Short: "List example sentences for a word",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			word := args[0]
			n := a.effectiveLimit(0)
			a.progressf("fetching examples for %q...", word)
			examples, err := a.client.Examples(cmd.Context(), a.lang, word, n)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(examples, len(examples))
		},
	}
}

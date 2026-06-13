package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

func (a *App) defineCmd() *cobra.Command {
	var pos string
	cmd := &cobra.Command{
		Use:   "define <word>",
		Short: "Show definitions for a word",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			word := args[0]
			n := a.effectiveLimit(0)
			a.progressf("looking up %q...", word)
			defs, err := a.client.Define(cmd.Context(), a.lang, word, n)
			if err != nil {
				return mapFetchErr(err)
			}
			if pos != "" {
				filtered := defs[:0]
				for _, d := range defs {
					if strings.EqualFold(d.PartOfSpeech, pos) {
						filtered = append(filtered, d)
					}
				}
				defs = filtered
			}
			return a.renderOrEmpty(defs, len(defs))
		},
	}
	cmd.Flags().StringVar(&pos, "pos", "", "filter by part-of-speech (e.g. noun, verb)")
	return cmd
}

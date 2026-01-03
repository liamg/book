package cli

import (
	"strings"

	"github.com/liamg/book/internal/bot"
	"github.com/spf13/cobra"
)

var sourceAddress string
var sourcePort int
var sourceChannel string
var useUndernet bool
var limitExtensions []string

var rootCmd = &cobra.Command{
	Use:   "book [query]",
	Short: "Book is an ebook search and download tool",
	Long:  `Book is a command-line tool to search and download ebooks from IRC channels.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
		if useUndernet {
			sourceAddress = "irc.undernet.org"
			sourceChannel = "#bookz"
			sourcePort = 6667
		}
		cfg := Config{
			Source: bot.Source{
				Address: sourceAddress,
				Port:    sourcePort,
				Channel: sourceChannel,
			},
			Query:           strings.Join(args, " "),
			LimitExtensions: limitExtensions,
		}
		return Run(cfg)
	},
}

func Execute() error {
	rootCmd.Flags().StringVarP(&sourceAddress, "server", "s", "irc.irchighway.net", "IRC server address")
	rootCmd.Flags().IntVarP(&sourcePort, "port", "p", 6667, "IRC server port")
	rootCmd.Flags().StringVarP(&sourceChannel, "channel", "c", "#ebooks", "IRC channel to join")
	rootCmd.Flags().BoolVarP(&useUndernet, "undernet", "u", false, "Use Undernet IRC network")
	rootCmd.Flags().StringSliceVarP(&limitExtensions, "ext", "e", []string{}, "Limit search to specific file extension(s) e.g. epub")
	return rootCmd.Execute()
}

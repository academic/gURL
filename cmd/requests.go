package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

var cmdGet = &cobra.Command{
	Use:   "GET [<url> to send GET request]",
	Short: "Fetches data from given url",
	Long: `Fetches data from given url.
  It should be valid URL.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("URL: " + strings.Join(args, " "))
	},
}

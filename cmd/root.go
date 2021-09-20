package cmd

import (
	"fmt"
	"github.com/academic/gURL/src"
	"github.com/spf13/cobra"
	"os"
)

var (
	url = ""
	proxy = ""
	c     = src.NewClient()
)

var rootCmd = &cobra.Command{
	Use:   "gURL [options...] <url>",
	Short: "gURL is open source CLI tool written in Go.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url = args[0]
		fmt.Println("URL: ", url)
		checkFlags()
	},
}

func Execute() {
	rootCmd.AddCommand(cmdGet)
	rootCmd.PersistentFlags().StringVarP(&proxy, "proxy", "x", "", "[protocol://]host[:port] Use this proxy")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// checkFlags checks the flags, get input, and sets the inputs to Client.
func checkFlags() {
	if proxy != "" {
		c.SetProxy(proxy)
	}
}

package cmd

import (
	"fmt"
	"github.com/academic/gURL/src"
	"github.com/spf13/cobra"
	"os"
)

var (
	URL   = ""
	proxy = ""
	c     = src.NewClient()
)

var rootCmd = &cobra.Command{
	Use:   "gURL [options...] <url>",
	Short: "gURL is open source CLI tool written in Go.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		URL = args[0]
		fmt.Println("URL: ", URL)
		err := checkFlags()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
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
func checkFlags() error {
	if proxy != "" {
		proxy, err := proxyCmd(proxy)
		if err != nil {
			return err
		}
		c.SetProxy(proxy)
	}
	return nil
}

package cmd

import (
	"fmt"
	"github.com/academic/gURL/src"
	"github.com/spf13/cobra"
	"os"
)

var (
	// URL is the target address.
	URL = ""

	// proxy is the url in the format of [protocol://]host[:port]. Its flag is
	// --proxy [protocol://]host[:port] or -x [protocol://]host[:port].
	proxy = ""

	// proxyUser is the username-password pair in <user:password> format. Its flag is
	// --proxy-user <user:password>.
	proxyUser = ""

	// proxyBasic is the flag variable whether indicates command contains --proxy-basic flag.
	// Basic is also the default unless anything else is asked for.
	proxyBasic = true

	// proxyDigest is the flag variable whether indicates command contains --proxy-digest flag.
	proxyDigest = false

	// proxyNTLM is the flag variable whether indicates command contains --proxy-ntlm flag.
	proxyNTLM = false

	// proxyNegotiate is the flag variable whether indicates command contains --proxy-negotiate flag.
	proxyNegotiate = false

	// c is the client.
	c = src.NewClient()
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
	rootCmd.PersistentFlags().StringVarP(&proxyUser, "proxy-user", "U", "", "<user:password> Proxy user and password")
	rootCmd.PersistentFlags().BoolVarP(&proxyBasic, "proxy-basic", "", true, "Use Basic authentication on the proxy")
	rootCmd.PersistentFlags().BoolVarP(&proxyDigest, "proxy-digest", "", false, "Use Digest authentication on the proxy")
	rootCmd.PersistentFlags().BoolVarP(&proxyNTLM, "proxy-ntlm", "", false, "Use NTLM authentication on the proxy")
	rootCmd.PersistentFlags().BoolVarP(&proxyNegotiate, "proxy-negotiate", "", false, "Use HTTP Negotiate (SPNEGO) authentication on the proxy")
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
	if proxyUser != "" {
		proxyUserCredentials, err := proxyUserCmd(proxyUser)
		if err != nil {
			return err
		}
		if !proxyNTLM && !proxyNegotiate && !proxyDigest { // Basic authentication
			c.AddHeader("Proxy-Authenticate", fmt.Sprintf("Basic %s", proxyUserCredentials))
		}
	}
	return nil
}

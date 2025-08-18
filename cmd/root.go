package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/academic/gURL/src"
	"github.com/spf13/cobra"
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

	// cookie Pass the data to the HTTP server in the Cookie header.
	// -b, --cookie <data|filename>
	cookieFile = ""
	cookies    []string

	// c is the client.
	c = src.NewClient()
)

var rootCmd = &cobra.Command{
	Use:   "gURL [options...] <url>",
	Short: "gURL is a command line tool for transferring data with URLs",
	Long: `gURL is a command line tool for transferring data from or to a server.
It supports various protocols and aims to be a curl-like tool written in Go.

Examples:
  gURL GET https://example.com
  gURL POST https://example.com -d "data=value"
  gURL GET https://example.com -H "Authorization: Bearer token"
  gURL -X PATCH https://example.com --json '{"key":"value"}'
  gURL POST https://example.com -T file.txt`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("URL is not provided")
			os.Exit(1)
		}
		URL = args[0]

		err := checkFlags()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// If method is specified with -X flag, use it
		httpMethod := "GET" // default
		if method != "" {
			httpMethod = strings.ToUpper(method)
		}

		executeRequest(httpMethod, URL)
	},
}

func Execute() {
	// Add all HTTP method commands
	rootCmd.AddCommand(cmdGet)
	rootCmd.AddCommand(cmdPost)
	rootCmd.AddCommand(cmdPut)
	rootCmd.AddCommand(cmdDelete)
	rootCmd.AddCommand(cmdHead)
	rootCmd.AddCommand(cmdOptions)
	rootCmd.AddCommand(cmdPatch)

	// Proxy flags
	rootCmd.PersistentFlags().StringVarP(&proxy, "proxy", "x", "", "[protocol://]host[:port] Use this proxy")
	rootCmd.PersistentFlags().StringVarP(&proxyUser, "proxy-user", "U", "", "<user:password> Proxy user and password")
	rootCmd.PersistentFlags().BoolVarP(&proxyBasic, "proxy-basic", "", true, "Use Basic authentication on the proxy")
	rootCmd.PersistentFlags().BoolVarP(&proxyDigest, "proxy-digest", "", false, "Use Digest authentication on the proxy")
	rootCmd.PersistentFlags().BoolVarP(&proxyNTLM, "proxy-ntlm", "", false, "Use NTLM authentication on the proxy")
	rootCmd.PersistentFlags().BoolVarP(&proxyNegotiate, "proxy-negotiate", "", false, "Use HTTP Negotiate (SPNEGO) authentication on the proxy")

	// Cookie flags
	rootCmd.PersistentFlags().StringSliceVarP(&cookies, "cookie", "b", []string{}, "Pass the data to the HTTP server in the Cookie header")

	// HTTP request flags
	rootCmd.PersistentFlags().StringSliceVarP(&headers, "header", "H", []string{}, "Pass custom header(s) to server")
	rootCmd.PersistentFlags().StringSliceVarP(&data, "data", "d", []string{}, "HTTP POST data")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "Write to file instead of stdout")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Make the operation more talkative")
	rootCmd.PersistentFlags().BoolVarP(&silent, "silent", "s", false, "Silent mode")
	rootCmd.PersistentFlags().BoolVarP(&includeHeaders, "include", "i", false, "Include protocol response headers in the output")
	rootCmd.PersistentFlags().BoolVarP(&followRedirects, "location", "L", false, "Follow redirects")
	rootCmd.PersistentFlags().IntVarP(&maxRedirects, "max-redirs", "", 50, "Maximum number of redirects allowed")
	rootCmd.PersistentFlags().IntVarP(&timeout, "max-time", "m", 0, "Maximum time allowed for the transfer")
	rootCmd.PersistentFlags().StringVarP(&userAgent, "user-agent", "A", "", "Send User-Agent <name> to server")
	rootCmd.PersistentFlags().StringVarP(&referer, "referer", "e", "", "Referrer URL")
	rootCmd.PersistentFlags().BoolVarP(&insecure, "insecure", "k", false, "Allow insecure server connections when using SSL")
	rootCmd.PersistentFlags().StringVarP(&method, "request", "X", "", "Specify request command to use")
	rootCmd.PersistentFlags().StringVarP(&uploadFile, "upload-file", "T", "", "Transfer local FILE to destination")
	rootCmd.PersistentFlags().StringSliceVarP(&formData, "form", "F", []string{}, "Specify multipart MIME data")
	rootCmd.PersistentFlags().StringVar(&jsonData, "json", "", "HTTP POST JSON data")
	rootCmd.PersistentFlags().StringVar(&rawData, "raw", "", "HTTP POST raw data")

	// HTTP version flags
	rootCmd.PersistentFlags().BoolVar(&http10, "http1.0", false, "Use HTTP 1.0")
	rootCmd.PersistentFlags().BoolVar(&http11, "http1.1", false, "Use HTTP 1.1")
	rootCmd.PersistentFlags().BoolVar(&http2, "http2", false, "Use HTTP 2")
	rootCmd.PersistentFlags().BoolVar(&http3, "http3", false, "Use HTTP 3")

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
	if len(cookies) > 0 {
		cookieMap := make(map[string]string)
		for _, cookie := range cookies {
			if strings.Contains(cookie, "=") {
				// Handle multiple cookies separated by &
				if strings.Contains(cookie, "&") {
					pairs := strings.Split(cookie, "&")
					for _, pair := range pairs {
						if cookieParts := strings.SplitN(pair, "=", 2); len(cookieParts) == 2 {
							cookieMap[strings.TrimSpace(cookieParts[0])] = strings.TrimSpace(cookieParts[1])
						}
					}
				} else {
					// Single cookie
					if cookieParts := strings.SplitN(cookie, "=", 2); len(cookieParts) == 2 {
						cookieMap[strings.TrimSpace(cookieParts[0])] = strings.TrimSpace(cookieParts[1])
					}
				}
			}
		}
		if len(cookieMap) > 0 {
			c.AddCookies(cookieMap)
		}
	}
	return nil
}

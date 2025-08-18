package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/academic/gURL/src"
	"github.com/spf13/cobra"
)

var (
	// Common flags for all HTTP methods
	headers         []string
	data            []string
	outputFile      string
	verbose         bool
	silent          bool
	includeHeaders  bool
	followRedirects bool
	maxRedirects    int
	timeout         int
	userAgent       string
	referer         string
	insecure        bool
	method          string
	uploadFile      string
	formData        []string
	jsonData        string
	rawData         string
)

var cmdGet = &cobra.Command{
	Use:   "GET [url]",
	Short: "Send GET request to the specified URL",
	Long:  `Send a GET request to the specified URL and display the response.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		URL = args[0]
		executeRequest("GET", URL)
	},
}

var cmdPost = &cobra.Command{
	Use:   "POST [url]",
	Short: "Send POST request to the specified URL",
	Long:  `Send a POST request to the specified URL with optional data.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		URL = args[0]
		executeRequest("POST", URL)
	},
}

var cmdPut = &cobra.Command{
	Use:   "PUT [url]",
	Short: "Send PUT request to the specified URL",
	Long:  `Send a PUT request to the specified URL with optional data.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		URL = args[0]
		executeRequest("PUT", URL)
	},
}

var cmdDelete = &cobra.Command{
	Use:   "DELETE [url]",
	Short: "Send DELETE request to the specified URL",
	Long:  `Send a DELETE request to the specified URL.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		URL = args[0]
		executeRequest("DELETE", URL)
	},
}

var cmdHead = &cobra.Command{
	Use:   "HEAD [url]",
	Short: "Send HEAD request to the specified URL",
	Long:  `Send a HEAD request to the specified URL (headers only).`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		URL = args[0]
		executeRequest("HEAD", URL)
	},
}

var cmdOptions = &cobra.Command{
	Use:   "OPTIONS [url]",
	Short: "Send OPTIONS request to the specified URL",
	Long:  `Send an OPTIONS request to the specified URL.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		URL = args[0]
		executeRequest("OPTIONS", URL)
	},
}

var cmdPatch = &cobra.Command{
	Use:   "PATCH [url]",
	Short: "Send PATCH request to the specified URL",
	Long:  `Send a PATCH request to the specified URL with optional data.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		URL = args[0]
		executeRequest("PATCH", URL)
	},
}

func executeRequest(httpMethod, url string) {
	// Apply timeout if specified
	if timeout > 0 {
		c.SetTimeout(time.Duration(timeout) * time.Second)
	}

	// Handle cookies
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

	// Add headers
	for _, header := range headers {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			c.AddHeader(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}

	// Add User-Agent if specified
	if userAgent != "" {
		c.AddHeader("User-Agent", userAgent)
	}

	// Add Referer if specified
	if referer != "" {
		c.AddHeader("Referer", referer)
	}

	// Handle data for POST/PUT/PATCH requests
	if len(data) > 0 {
		dataStr := strings.Join(data, "&")
		c.AddBodyBytes([]byte(dataStr))
		if httpMethod == "POST" || httpMethod == "PUT" || httpMethod == "PATCH" {
			c.AddHeader("Content-Type", "application/x-www-form-urlencoded")
		}
	}

	// Handle JSON data
	if jsonData != "" {
		c.AddBodyBytes([]byte(jsonData))
		c.AddHeader("Content-Type", "application/json")
	}

	// Handle raw data
	if rawData != "" {
		c.AddBodyBytes([]byte(rawData))
	}

	// Handle form data
	for _, form := range formData {
		parts := strings.SplitN(form, "=", 2)
		if len(parts) == 2 {
			c.AddParam(parts[0], parts[1])
		}
	}

	// Handle file upload
	if uploadFile != "" {
		c.AddFile("file", uploadFile)
	}

	// Execute the request
	var response *src.Response
	var err error

	// If file upload is specified, use SendFile method
	if uploadFile != "" {
		response, err = c.SendFile(url)
	} else {
		switch httpMethod {
		case "GET":
			response, err = c.Get(url)
		case "POST":
			response, err = c.Post(url)
		case "PUT":
			response, err = c.Put(url)
		case "DELETE":
			response, err = c.Delete(url)
		case "OPTIONS":
			response, err = c.Options(url)
		case "HEAD":
			response, err = c.Head(url)
		case "PATCH":
			response, err = c.Patch(url)
		default:
			// Use generic request method for any other HTTP method
			response, err = c.Request(httpMethod, url)
		}
	}

	if err != nil {
		if !silent {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}

	handleResponse(response, httpMethod)
}

func handleResponse(response *src.Response, httpMethod string) {
	var output io.Writer = os.Stdout

	// Handle output file
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			if !silent {
				fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			}
			os.Exit(1)
		}
		defer file.Close()
		output = file
	}

	// Show headers if requested or verbose
	if includeHeaders || verbose || httpMethod == "HEAD" {
		if verbose {
			fmt.Fprintf(output, "< HTTP/1.1 %d\n", response.StatusCode)
		}
		for key, value := range response.Header.Mapper {
			if verbose {
				fmt.Fprintf(output, "< %s: %s\n", key, value)
			} else if includeHeaders {
				fmt.Fprintf(output, "%s: %s\n", key, value)
			}
		}
		if includeHeaders || verbose {
			fmt.Fprintf(output, "\n")
		}
	}

	// Show body (unless it's a HEAD request)
	if httpMethod != "HEAD" && response.Body != nil && len(response.Body) > 0 {
		output.Write(response.Body)
		if outputFile == "" {
			fmt.Fprintf(output, "\n")
		}
	}

	// Show status code if verbose and not silent
	if verbose && !silent && outputFile == "" {
		fmt.Fprintf(os.Stderr, "HTTP Status: %d\n", response.StatusCode)
	}
}

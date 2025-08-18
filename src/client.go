package src

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/quic-go/quic-go/http3"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"golang.org/x/net/http2"
)

var (
	defaultTimeDuration = time.Second * 10

	defaultContentType = "application/x-www-form-urlencoded"

	jsonContentType = "application/json"

	formContentType = "multipart/form-data"

	ErrEmptyURL  = errors.New("empty url")
	ErrEmptyFile = errors.New("empty file")

	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

type Client struct {
	proxy       string // set to all requests
	timeout     time.Duration
	crt         *tls.Certificate
	opts        *requestOptions
	httpVersion string // "1.0", "1.1", "2", "3"
	insecure    bool   // allow insecure SSL
	// Authentication fields
	authType string // "basic", "digest", "ntlm", "negotiate"
	username string
	password string
}

func NewClientPool() sync.Pool {
	return sync.Pool{
		New: func() interface{} {
			return &Client{
				timeout: defaultTimeDuration,
				crt:     nil,
				opts:    newRequestOptions(),
			}
		},
	}
}

func NewClient() *Client {
	return &Client{
		timeout:     defaultTimeDuration,
		crt:         nil,
		opts:        newRequestOptions(),
		httpVersion: "1.1", // default to HTTP/1.1
		insecure:    false,
	}
}

func (c *Client) SetProxy(proxy string) *Client {
	c.proxy = proxy
	return c
}

func (c *Client) SetTimeout(duration time.Duration) *Client {
	c.timeout = duration
	return c
}

func (c *Client) SetHTTPVersion(version string) *Client {
	c.httpVersion = version
	return c
}

func (c *Client) SetInsecure(insecure bool) *Client {
	c.insecure = insecure
	return c
}

func (c *Client) SetCrt(certPath, keyPath string) *Client {
	clientCrt, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		// todo handle this err
		clientCrt = tls.Certificate{}
	}
	c.crt = &clientCrt
	return c
}

// SetBasicAuth configures HTTP Basic Authentication
func (c *Client) SetBasicAuth(userPass string) *Client {
	parts := strings.SplitN(userPass, ":", 2)
	if len(parts) == 2 {
		c.authType = "basic"
		c.username = parts[0]
		c.password = parts[1]
	}
	return c
}

// SetDigestAuth configures HTTP Digest Authentication
func (c *Client) SetDigestAuth(userPass string) *Client {
	parts := strings.SplitN(userPass, ":", 2)
	if len(parts) == 2 {
		c.authType = "digest"
		c.username = parts[0]
		c.password = parts[1]
	}
	return c
}

// SetNTLMAuth configures HTTP NTLM Authentication
func (c *Client) SetNTLMAuth(userPass string) *Client {
	parts := strings.SplitN(userPass, ":", 2)
	if len(parts) == 2 {
		c.authType = "ntlm"
		c.username = parts[0]
		c.password = parts[1]
	}
	return c
}

// SetNegotiateAuth configures HTTP Negotiate (SPNEGO) Authentication
func (c *Client) SetNegotiateAuth(userPass string) *Client {
	parts := strings.SplitN(userPass, ":", 2)
	if len(parts) == 2 {
		c.authType = "negotiate"
		c.username = parts[0]
		c.password = parts[1]
	}
	return c
}

// addAuthenticationHeaders adds authentication headers based on the configured auth type
func (c *Client) addAuthenticationHeaders(headers requestHeaders) {
	switch c.authType {
	case "basic":
		auth := base64.StdEncoding.EncodeToString([]byte(c.username + ":" + c.password))
		headers.normal.Set("Authorization", "Basic "+auth)
	case "digest":
		// For digest auth, we need to handle the initial request and response challenge
		// This is a simplified implementation - in practice, digest auth requires
		// parsing the WWW-Authenticate header from a 401 response
		headers.normal.Set("Authorization", "Digest username=\""+c.username+"\"")
	case "ntlm":
		// NTLM requires a complex handshake - this is a placeholder
		headers.normal.Set("Authorization", "NTLM")
	case "negotiate":
		// Negotiate (SPNEGO) requires GSSAPI/Kerberos - this is a placeholder
		headers.normal.Set("Authorization", "Negotiate")
	}
}

// parseDigestChallenge parses a digest challenge and returns authentication header
func (c *Client) parseDigestChallenge(challenge, method, uri string) string {
	// Parse the challenge parameters
	params := make(map[string]string)

	// Simple parsing - in production this should be more robust
	parts := strings.Split(challenge, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if idx := strings.Index(part, "="); idx > 0 {
			key := strings.TrimSpace(part[:idx])
			value := strings.Trim(strings.TrimSpace(part[idx+1:]), "\"")
			params[key] = value
		}
	}

	realm := params["realm"]
	nonce := params["nonce"]
	qop := params["qop"]

	// Generate response hash
	ha1 := fmt.Sprintf("%x", md5.Sum([]byte(c.username+":"+realm+":"+c.password)))
	ha2 := fmt.Sprintf("%x", md5.Sum([]byte(method+":"+uri)))

	var response string
	if qop == "auth" {
		nc := "00000001"
		cnonce := "0a4f113b"
		response = fmt.Sprintf("%x", md5.Sum([]byte(ha1+":"+nonce+":"+nc+":"+cnonce+":"+qop+":"+ha2)))
		return fmt.Sprintf("Digest username=\"%s\", realm=\"%s\", nonce=\"%s\", uri=\"%s\", qop=%s, nc=%s, cnonce=\"%s\", response=\"%s\"",
			c.username, realm, nonce, uri, qop, nc, cnonce, response)
	} else {
		response = fmt.Sprintf("%x", md5.Sum([]byte(ha1+":"+nonce+":"+ha2)))
		return fmt.Sprintf("Digest username=\"%s\", realm=\"%s\", nonce=\"%s\", uri=\"%s\", response=\"%s\"",
			c.username, realm, nonce, uri, response)
	}
}

func (c *Client) AddParam(key, value string) *Client {
	c.opts.params.Set(key, value)
	return c
}

func (c *Client) AddParams(params Mapper) *Client {
	for key, value := range params {
		c.opts.params.Set(key, value)
	}
	return c
}

func (c *Client) AddHeader(key, value string) *Client {
	c.opts.headers.normal.Set(key, value)
	return c
}

func (c *Client) AddHeaders(headers Mapper) *Client {
	for key, value := range headers {
		c.opts.headers.normal.Set(key, value)
	}
	return c
}

func (c *Client) AddCookie(key, value string) *Client {
	c.opts.headers.cookies.Set(key, value)
	return c
}

func (c *Client) AddCookies(cookies Mapper) *Client {
	for key, value := range cookies {
		c.opts.headers.cookies.Set(key, value)
	}
	return c
}

func (c *Client) AddFile(fileName, filePath string) *Client {
	c.opts.files.Set(fileName, filePath)
	return c
}

func (c *Client) AddFiles(files Mapper) *Client {
	for key, value := range files {
		c.opts.files.Set(key, value)
	}
	return c
}

func (c *Client) AddBodyByte(body []byte) *Client {
	c.opts.body = body
	return c
}

func (c *Client) AddBodyStruct(object interface{}) *Client {
	bodyByte, _ := json.Marshal(object)
	c.opts.body = bodyByte
	return c
}

func (c *Client) AddBodyBytes(bodyBytes []byte) *Client {
	c.opts.body = bodyBytes
	return c
}

func (c *Client) Get(rawUrl string) (*Response, error) {
	if rawUrl == "" {
		return nil, ErrEmptyURL
	}
	var (
		urlValue = url.Values{}
		err      error
	)
	queryArray := strings.SplitN(rawUrl, "?", 2)
	if len(queryArray) != 1 {
		urlValue, err = url.ParseQuery(queryArray[1])
		if err != nil {
			return nil, err
		}
	}
	for key, value := range c.opts.params.Mapper {
		urlValue.Set(key, value)
	}
	reqUrl := addString(queryArray[0], "?", urlValue.Encode())
	return c.call(reqUrl, fasthttp.MethodGet, c.opts.headers, nil)
}

func (c *Client) Post(url string) (*Response, error) {
	if url == "" {
		return nil, ErrEmptyURL
	}

	return c.call(url, fasthttp.MethodPost, c.opts.headers, c.opts.body)
}

func (c *Client) Put(url string) (*Response, error) {
	if url == "" {
		return nil, ErrEmptyURL
	}

	return c.call(url, fasthttp.MethodPut, c.opts.headers, c.opts.body)
}

func (c *Client) Delete(url string) (*Response, error) {
	if url == "" {
		return nil, ErrEmptyURL
	}

	return c.call(url, fasthttp.MethodDelete, c.opts.headers, c.opts.body)
}

func (c *Client) Options(url string) (*Response, error) {
	if url == "" {
		return nil, ErrEmptyURL
	}

	return c.call(url, fasthttp.MethodOptions, c.opts.headers, c.opts.body)
}

func (c *Client) Head(url string) (*Response, error) {
	if url == "" {
		return nil, ErrEmptyURL
	}

	return c.call(url, fasthttp.MethodHead, c.opts.headers, nil)
}

func (c *Client) Patch(url string) (*Response, error) {
	if url == "" {
		return nil, ErrEmptyURL
	}

	return c.call(url, fasthttp.MethodPatch, c.opts.headers, c.opts.body)
}

// Request allows any HTTP method
func (c *Client) Request(method, url string) (*Response, error) {
	if url == "" {
		return nil, ErrEmptyURL
	}

	return c.call(url, method, c.opts.headers, c.opts.body)
}

func (c *Client) SendFile(url string) (*Response, error) {
	if url == "" {
		return nil, ErrEmptyURL
	}
	if len(c.opts.files.Mapper) == 0 {
		return nil, ErrEmptyFile
	}
	bodyBuffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuffer)
	for fileName, filePath := range c.opts.files.Mapper {
		fileWriter, err := bodyWriter.CreateFormFile(fileName, path.Base(filePath))
		if err != nil {
			return nil, err
		}

		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(fileWriter, file)
		if err != nil {
			_ = file.Close()
			return nil, err
		}
		_ = file.Close()
	}
	_ = bodyWriter.Close()
	c.opts.headers.normal.Set("content-type", bodyWriter.FormDataContentType())

	return c.call(url, fasthttp.MethodPost, c.opts.headers, bodyBuffer.Bytes())
}

func (c *Client) call(url, method string, headers requestHeaders, body []byte) (*Response, error) {
	// Use HTTP/2 or HTTP/3 if specified
	if c.httpVersion == "2" || c.httpVersion == "3" {
		return c.callHTTP2OrHTTP3(url, method, headers, body)
	}

	// Use fasthttp for HTTP/1.x
	return c.callFastHTTP(url, method, headers, body)
}

func (c *Client) callFastHTTP(url, method string, headers requestHeaders, body []byte) (*Response, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.SetMethod(method)

	// Add authentication headers if configured
	c.addAuthenticationHeaders(headers)

	// Handle cookies by creating a proper Cookie header
	if len(headers.cookies.Mapper) > 0 {
		var cookiePairs []string
		for key, value := range headers.cookies.Mapper {
			cookiePairs = append(cookiePairs, fmt.Sprintf("%s=%s", key, value))
		}
		if len(cookiePairs) > 0 {
			req.Header.Set("Cookie", strings.Join(cookiePairs, "; "))
		}
	}

	for key, value := range headers.normal.Mapper {
		req.Header.Set(key, value)
	}

	if !req.Header.IsGet() && !req.Header.IsHead() {
		contentType := string(req.Header.ContentType())
		switch contentType {
		case jsonContentType:
			if body != nil {
				req.SetBody(body)
			}
		case defaultContentType:
			// For form data, just set the body directly
			if body != nil {
				req.SetBody(body)
			}
		default:
			if strings.Contains(contentType, formContentType) {
				// For multipart form data, set body directly
				if body != nil {
					req.SetBody(body)
				}
			} else if contentType == "" && body != nil {
				// If no content type is set and we have body data,
				// check if it looks like form data or JSON
				bodyStr := string(body)
				if strings.Contains(bodyStr, "=") && strings.Contains(bodyStr, "&") {
					// Looks like form data
					req.Header.SetContentType(defaultContentType)
					req.SetBody(body)
				} else if (strings.HasPrefix(bodyStr, "{") && strings.HasSuffix(bodyStr, "}")) ||
					(strings.HasPrefix(bodyStr, "[") && strings.HasSuffix(bodyStr, "]")) {
					// Looks like JSON
					req.Header.SetContentType(jsonContentType)
					req.SetBody(body)
				} else {
					// Just set the body as-is
					req.SetBody(body)
				}
			} else {
				req.SetBody(body)
			}
		}
	}

	client := &fasthttp.Client{
		ReadTimeout: c.timeout,
	}
	if c.crt != nil {
		client.TLSConfig = &tls.Config{
			InsecureSkipVerify: c.insecure,
			Certificates:       []tls.Certificate{*c.crt},
		}
	} else if c.insecure {
		client.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	if c.proxy != "" {
		client.Dial = fasthttpproxy.FasthttpHTTPDialer(c.proxy)
	}

	if err := client.Do(req, resp); err != nil {
		return nil, err
	}

	ret := &Response{
		Cookie:     RequestCookies{Mapper: NewCookies()},
		Header:     RequestHeaders{Mapper: NewHeaders()},
		StatusCode: resp.StatusCode(),
		Body:       resp.Body(),
	}
	resp.Header.VisitAll(func(key, value []byte) {
		ret.Header.Set(string(key), string(value))
	})
	resp.Header.VisitAllCookie(func(key, value []byte) {
		ret.Cookie.Set(string(key), string(value))
	})
	return ret, nil
}

func (c *Client) callHTTP2OrHTTP3(url, method string, headers requestHeaders, body []byte) (*Response, error) {
	var client *http.Client

	if c.httpVersion == "3" {
		// HTTP/3 client
		client = &http.Client{
			Transport: &http3.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: c.insecure,
				},
			},
			Timeout: c.timeout,
		}
	} else {
		// HTTP/2 client
		tlsConfig := &tls.Config{
			InsecureSkipVerify: c.insecure,
		}
		if c.crt != nil {
			tlsConfig.Certificates = []tls.Certificate{*c.crt}
		}

		transport := &http2.Transport{
			TLSClientConfig: tlsConfig,
		}

		client = &http.Client{
			Transport: transport,
			Timeout:   c.timeout,
		}
	}

	// Create request
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	// Add authentication headers if configured
	c.addAuthenticationHeaders(headers)

	// Set headers
	for key, value := range headers.normal.Mapper {
		req.Header.Set(key, value)
	}

	// Handle cookies
	if len(headers.cookies.Mapper) > 0 {
		var cookiePairs []string
		for key, value := range headers.cookies.Mapper {
			cookiePairs = append(cookiePairs, fmt.Sprintf("%s=%s", key, value))
		}
		if len(cookiePairs) > 0 {
			req.Header.Set("Cookie", strings.Join(cookiePairs, "; "))
		}
	}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Convert response
	ret := &Response{
		Cookie:     RequestCookies{Mapper: NewCookies()},
		Header:     RequestHeaders{Mapper: NewHeaders()},
		StatusCode: resp.StatusCode,
		Body:       respBody,
	}

	// Copy headers
	for key, values := range resp.Header {
		if len(values) > 0 {
			ret.Header.Set(key, values[0])
		}
	}

	// Copy cookies
	for _, cookie := range resp.Cookies() {
		ret.Cookie.Set(cookie.Name, cookie.Value)
	}

	return ret, nil
}

type Response struct {
	StatusCode int
	Body       []byte
	Header     RequestHeaders
	Cookie     RequestCookies
}

func addString(ss ...string) string {
	b := strings.Builder{}
	for _, s := range ss {
		b.WriteString(s)
	}
	return b.String()
}

type (
	Mapper map[string]string

	RequestParams  struct{ Mapper }
	RequestHeaders struct{ Mapper }
	RequestCookies struct{ Mapper }
	RequestFiles   struct{ Mapper }
)

func NewParams() Mapper {
	return make(map[string]string)
}

func NewHeaders() Mapper {
	return make(map[string]string)
}

func NewCookies() Mapper {
	return make(map[string]string)
}

func NewFiles() Mapper {
	return make(map[string]string)
}

func (m Mapper) Get(key string) string {
	value, ok := m[key]
	if ok {
		return value
	}
	return ""
}

func (m Mapper) Set(key, value string) Mapper {
	m[key] = value
	return m
}

func newRequestOptions() *requestOptions {
	return &requestOptions{
		files: RequestFiles{Mapper: NewFiles()},
		headers: requestHeaders{
			normal:  RequestHeaders{Mapper: NewHeaders()},
			cookies: RequestCookies{Mapper: NewCookies()},
		},
		params: RequestParams{Mapper: NewParams()},
	}
}

type requestOptions struct {
	body    []byte
	Proxy   string
	files   RequestFiles
	headers requestHeaders
	params  RequestParams
}

type requestHeaders struct {
	normal  RequestHeaders
	cookies RequestCookies
}

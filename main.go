package main

import (
	"bytes"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
)

var (
	defaultTimeDuration = time.Second * 10

	defaultContentType = "application/x-www-form-urlencoded"

	jsonContentType = "application/json"

	formContentType = "multipart/form-data"

	EmptyUrlErr  = errors.New("empty url")
	EmptyFileErr = errors.New("empty file")

	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

type Client struct {
	proxy   string // set to all requests
	timeout time.Duration
	crt     *tls.Certificate
	opts    *requestOptions
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
		timeout: defaultTimeDuration,
		crt:     nil,
		opts:    newRequestOptions(),
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

func (c *Client) SetCrt(certPath, keyPath string) *Client {
	clientCrt, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		// todo handle this err
		clientCrt = tls.Certificate{}
	}
	c.crt = &clientCrt
	return c
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
		return nil, EmptyUrlErr
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
		return nil, EmptyUrlErr
	}

	return c.call(url, fasthttp.MethodPost, c.opts.headers, c.opts.body)
}

func (c *Client) Put(url string) (*Response, error) {
	if url == "" {
		return nil, EmptyUrlErr
	}

	return c.call(url, fasthttp.MethodPut, c.opts.headers, c.opts.body)
}

func (c *Client) Delete(url string) (*Response, error) {
	if url == "" {
		return nil, EmptyUrlErr
	}

	return c.call(url, fasthttp.MethodDelete, c.opts.headers, c.opts.body)
}

func (c *Client) Options(url string) (*Response, error) {
	if url == "" {
		return nil, EmptyUrlErr
	}

	return c.call(url, fasthttp.MethodOptions, c.opts.headers, c.opts.body)
}

func (c *Client) SendFile(url string, options ...RequestOption) (*Response, error) {
	if url == "" {
		return nil, EmptyUrlErr
	}
	if len(c.opts.files.Mapper) == 0 {
		return nil, EmptyFileErr
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
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.SetMethod(method)

	for key, value := range headers.cookies.Mapper {
		req.Header.SetCookie(key, value)
	}

	for key, value := range headers.normal.Mapper {
		req.Header.Set(key, value)
	}

	if !req.Header.IsGet() {
		contentType := string(req.Header.ContentType())
		switch contentType {
		case jsonContentType:
			if body != nil {
				req.SetBody(body)
			}
		default:
			if !strings.Contains(contentType, formContentType) && body != nil {
				argsMap := make(map[string]interface{})
				if err := json.Unmarshal(body, &argsMap); err != nil {
					return nil, err
				}
				fastArgs := new(fasthttp.Args)
				for key, value := range argsMap {
					fastArgs.Add(key, fmt.Sprintf("%v", value))
				}
				req.SetBody(fastArgs.QueryString())
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
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{*c.crt},
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

type RequestOption struct {
	f func(*requestOptions)
}

func main2() {

	flag.Parse()

	res, err := NewClient().
		AddParam("param1", "param1").
		AddHeader("header1", "header1").
		AddCookie("cookie1", "cookie1").
		Get("https://example.com")
	if err == nil {
		fmt.Println(string(res.Body))
	}

}

func main() {
	var echoTimes int

	var cmdPrint = &cobra.Command{
		Use:   "print [string to print]",
		Short: "Print anything to the screen",
		Long: `print is for printing anything back to the screen.
  For many years people have printed back to the screen.`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Print: " + strings.Join(args, " "))
		},
	}

	var cmdEcho = &cobra.Command{
		Use:   "echo [string to echo]",
		Short: "Echo anything to the screen",
		Long: `echo is for echoing anything back.
  Echo works a lot like print, except it has a child command.`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Echo: " + strings.Join(args, " "))
		},
	}

	var cmdTimes = &cobra.Command{
		Use:   "times [string to echo]",
		Short: "Echo anything to the screen more times",
		Long: `echo things multiple times back to the user by providing
  a count and a string.`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			for i := 0; i < echoTimes; i++ {
				fmt.Println("Echo: " + strings.Join(args, " "))
			}
		},
	}

	cmdTimes.Flags().IntVarP(&echoTimes, "times", "t", 1, "times to echo the input")

	var rootCmd = &cobra.Command{Use: "gurl"}
	rootCmd.AddCommand(cmdPrint, cmdEcho)
	cmdEcho.AddCommand(cmdTimes)
	rootCmd.Execute()
}

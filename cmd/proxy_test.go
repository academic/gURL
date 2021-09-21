package cmd

import (
	"strings"
	"testing"
)

const (
	proxyUrl            = "proxy://myproxy:1234"
	httpProxyUrl        = "http://myproxy:1234"
	missingProxyUrl     = "proxy://proxy"
	missingHTTPProxyUrl = "http://myproxy"
)

func TestProxyCmd_ProxyUrl(t *testing.T) {
	testProxyUrl, err := proxyCmd(proxyUrl)
	if err != nil {
		t.Errorf(err.Error())
	}
	if !strings.HasSuffix(testProxyUrl, proxyUrl) {
		t.Errorf("wrong proxy url. expected url: %s, got: %s", proxyUrl, testProxyUrl)
	}
}

func TestProxyCmd_HTTPUrl(t *testing.T) {
	testProxyUrl, err := proxyCmd(httpProxyUrl)
	if err != nil {
		t.Errorf(err.Error())
	}
	if !strings.HasSuffix(testProxyUrl, httpProxyUrl) {
		t.Errorf("wrong proxy url. expected url: %s, got: %s", httpProxyUrl, testProxyUrl)
	}
}

func TestProxyCmd_MissingProxyUrl(t *testing.T) {
	expectedUrl := missingProxyUrl + ":1080"
	testProxyUrl, err := proxyCmd(missingProxyUrl)
	if err != nil {
		t.Errorf(err.Error())
	}
	if !strings.HasSuffix(testProxyUrl, expectedUrl) {
		t.Errorf("wrong proxy url. expected url: %s, got: %s", expectedUrl, testProxyUrl)
	}
}

func TestProxyCmd_MissingHTTPProxyUrl(t *testing.T) {
	expectedUrl := missingHTTPProxyUrl + ":1080"
	testProxyUrl, err := proxyCmd(missingHTTPProxyUrl)
	if err != nil {
		t.Errorf(err.Error())
	}
	if !strings.HasSuffix(testProxyUrl, expectedUrl) {
		t.Errorf("wrong proxy url. expected url: %s, got: %s", expectedUrl, testProxyUrl)
	}
}

package cmd

import (
	"log"
	"testing"
)

const (
	proxyUrl             = "proxy://myproxy:1234"
	httpProxyUrl         = "http://myproxy:1234"
	missingProxyUrl      = "proxy://proxy"
	missingHTTPProxyUrl  = "http://myproxy"
	validProxyUser       = "user:pass"
	emptyPassProxyUser   = "user:"
	emptyUserProxyUser   = ":pass"
	invalidProxyUserLong = "user:pass1:pass2"
)

func TestProxyCmd_ProxyUrl(t *testing.T) {
	testProxyUrl, err := proxyCmd(proxyUrl)
	if err != nil {
		t.Errorf(err.Error())
	}
	if testProxyUrl != proxyUrl {
		t.Errorf("wrong proxy url. expected url: %s, got: %s", proxyUrl, testProxyUrl)
	}
}

func TestProxyCmd_HTTPUrl(t *testing.T) {
	testProxyUrl, err := proxyCmd(httpProxyUrl)
	if err != nil {
		t.Errorf(err.Error())
	}
	if testProxyUrl != httpProxyUrl {
		t.Errorf("wrong proxy url. expected url: %s, got: %s", httpProxyUrl, testProxyUrl)
	}
}

func TestProxyCmd_MissingProxyUrl(t *testing.T) {
	expectedUrl := missingProxyUrl + ":1080"
	testProxyUrl, err := proxyCmd(missingProxyUrl)
	if err != nil {
		t.Errorf(err.Error())
	}
	if testProxyUrl != expectedUrl {
		t.Errorf("wrong proxy url. expected url: %s, got: %s", expectedUrl, testProxyUrl)
	}
}

func TestProxyCmd_MissingHTTPProxyUrl(t *testing.T) {
	expectedUrl := missingHTTPProxyUrl + ":1080"
	testProxyUrl, err := proxyCmd(missingHTTPProxyUrl)
	if err != nil {
		t.Errorf(err.Error())
	}
	if testProxyUrl != expectedUrl {
		t.Errorf("wrong proxy url. expected url: %s, got: %s", expectedUrl, testProxyUrl)
	}
}

func TestCheckValidProxyUser(t *testing.T) {
	err := checkProxyUser(validProxyUser)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestCheckEmptyUserProxyUser(t *testing.T) {
	err := checkProxyUser(emptyUserProxyUser)
	if err == nil {
		t.Errorf("empyt username.")
	}
}

func TestCheckEmptyPassProxyUser(t *testing.T) {
	err := checkProxyUser(emptyPassProxyUser)
	if err == nil {
		t.Errorf("invalid password.")
	}
}

func TestCheckEmptyProxyUser(t *testing.T) {
	err := checkProxyUser("")
	if err == nil {
		t.Errorf("empty user")
	}
}

func TestCheckLongProxyUser(t *testing.T) {
	err := checkProxyUser(invalidProxyUserLong)
	if err == nil {
		t.Errorf("long proxy user")
	}
}

func TestProxyUserCmdBasic(t *testing.T) {
	proxyUserCredential, err := proxyUserCmd(validProxyUser)
	if err != nil {
		t.Errorf(err.Error())
	}
	if proxyUserCredential == validProxyUser {
		t.Errorf("proxy user: <%s> should not be the same with <%s>", validProxyUser, proxyUserCredential)
	}
	log.Printf("proxy user credential: %s", proxyUserCredential)
}

func TestProxyUserCmdDigest(t *testing.T) {
	proxyDigest = true
	proxyUserCredential, err := proxyUserCmd(validProxyUser)
	if err != nil {
		t.Errorf(err.Error())
	}
	if proxyUserCredential != validProxyUser {
		t.Errorf("proxy user: <%s> should be the same with <%s>", validProxyUser, proxyUserCredential)
	}
}

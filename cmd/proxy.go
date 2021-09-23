package cmd

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
)

const defaultPort = "1080"

// proxyCmd parses proxy url and adds default port number if not exists.
func proxyCmd(proxy string) (string, error) {
	u, err := url.Parse(proxy)
	if err != nil {
		err = fmt.Errorf("not valid proxy address: %s", err.Error())
		return "", err
	}
	if u.Port() == "" {
		if proxy[len(proxy)-1] != ':' {
			proxy += ":"
		}
		proxy += defaultPort
	}
	return proxy, nil
}

// proxyUserCmd handles "--proxy-user" related tasks. Returns
// encoded with Base64 string.
func proxyUserCmd(proxyUser string) (string, error) {
	err := checkProxyUser(proxyUser)
	if err != nil {
		return "", err
	}
	if !proxyNTLM && !proxyNegotiate && !proxyDigest { // Basic auth
		encodedProxy := convertBasicAuth(proxyUser)
		return encodedProxy, nil
	}
	return proxyUser, nil
}

// checkProxyUser controls proxyUser whether is in <username:password> format.
func checkProxyUser(proxyUser string) error {
	p := strings.Split(proxyUser, ":")
	if len(p) != 2 || len(p[0]) == 0 || len(p[1]) == 0 {
		return fmt.Errorf("need to specify username and password in <username:password> format")
	}
	return nil
}

// convertBasicAuth encodes proxy username and password to Base64.
func convertBasicAuth(proxyUser string) string {
	return base64.StdEncoding.EncodeToString([]byte(proxyUser))
}

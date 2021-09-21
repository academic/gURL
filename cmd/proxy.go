package cmd

import (
	"fmt"
	"net/url"
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

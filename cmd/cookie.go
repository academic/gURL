package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

// cookieCMD Pass the data to the HTTP server in the Cookie header.
// -b, --cookie data
func cookiesCmd(cookiesCommandString []string) (map[string]string, error) {

	cookies := make(map[string]string)
	for _, cookieSlice := range cookiesCommandString {
		cookieSplit := strings.Split(cookieSlice, "=")
		cookies[cookieSplit[0]] = cookieSplit[1]
	}

	return cookies, nil
}

// cookieFileCmd Pass the data to the HTTP server in the Cookie header.
// -b, --cookie filename
func cookieFileCmd(cookieFileString string) (*http.Cookie, error) {

	if _, err := os.Stat(cookieFileString); err == nil {
		var cookie *http.Cookie
		file, err := os.Open(cookieFileString)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		lineNum := 1
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			cookieLine := scanner.Text()
			cookie, err = parseCookieLine(cookieLine, lineNum)
			if cookie == nil {
				return nil, err
			}
			if err != nil {
				return nil, err
			}
		}

		return cookie, nil

	} else {
		return nil, err

	}
}

// cookieJarCMD Specify to which file you want curl to write all cookies after a completed operation.
// -c, --cookie-jar <filename>
func cookieJarCmd(r http.Request, cookieJarDirectory string) (string, error) {
	cookieJarFile := cookieJarDirectory + "/" + r.URL.Host + ".txt"

	cookieJar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})

	if err != nil {
		return "", err
	}

	for _, c := range r.Cookies() {
		cookieJar.SetCookies(r.URL, []*http.Cookie{c})
	}

	jarFile, err := os.Create(cookieJarFile)
	if err != nil {
		return "", err
	}

	defer jarFile.Close()

	for _, cookies := range cookieJar.Cookies(r.URL) {
		cookieLine := cookies.Name + "=" + cookies.Value + "; " + "Path=" + cookies.Path + "; " + "Domain=" + cookies.Domain + "; " + "Expires=" + cookies.Expires.Format("Mon, 02 Jan 2006 15:04:05 MST") + "; " + "Secure=" + strconv.FormatBool(cookies.Secure) + "; " + "HttpOnly=" + strconv.FormatBool(cookies.HttpOnly) + "\n"
		jarFile.WriteString(cookieLine)
	}

	return "", nil
}

const httpOnlyPrefix = "#HttpOnly_"

func parseCookieLine(cookieLine string, lineNum int) (*http.Cookie, error) {
	var err error
	cookieLineHttpOnly := false
	if strings.HasPrefix(cookieLine, httpOnlyPrefix) {
		cookieLineHttpOnly = true
		cookieLine = strings.TrimPrefix(cookieLine, httpOnlyPrefix)
	}

	if strings.HasPrefix(cookieLine, "#") || cookieLine == "" {
		return nil, nil
	}

	cookieFields := strings.Split(cookieLine, "\t")

	if len(cookieFields) < 6 || len(cookieFields) > 7 {
		return nil, fmt.Errorf("incorrect number of fields in line %d.  Expected 6 or 7, got %d.", lineNum, len(cookieFields))
	}

	for i, v := range cookieFields {
		cookieFields[i] = strings.TrimSpace(v)
	}

	cookie := &http.Cookie{
		Domain:   cookieFields[0],
		Path:     cookieFields[2],
		Name:     cookieFields[5],
		HttpOnly: cookieLineHttpOnly,
	}
	cookie.Secure, err = strconv.ParseBool(cookieFields[3])
	if err != nil {
		return nil, err
	}
	expiresInt, err := strconv.ParseInt(cookieFields[4], 10, 64)
	if err != nil {
		return nil, err
	}
	if expiresInt > 0 {
		cookie.Expires = time.Unix(expiresInt, 0)
	}

	if len(cookieFields) == 7 {
		cookie.Value = cookieFields[6]
	}

	return cookie, nil
}

// LoadCookieJarFile takes a path to a curl (netscape) cookie jar file and crates a go http.CookieJar with the contents
func LoadCookieJarFile(path string) (http.CookieJar, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lineNum := 1
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		cookieLine := scanner.Text()
		cookie, err := parseCookieLine(cookieLine, lineNum)
		if cookie == nil {
			continue
		}
		if err != nil {
			return nil, err
		}

		var cookieScheme string
		if cookie.Secure {
			cookieScheme = "https"
		} else {
			cookieScheme = "http"
		}
		cookieUrl := &url.URL{
			Scheme: cookieScheme,
			Host:   cookie.Domain,
		}

		cookies := jar.Cookies(cookieUrl)
		cookies = append(cookies, cookie)
		jar.SetCookies(cookieUrl, cookies)

		lineNum++
	}

	return jar, nil
}

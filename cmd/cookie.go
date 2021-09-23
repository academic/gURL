package cmd

import (
	"fmt"
	"os"
)

// cookieCMD Pass the data to the HTTP server in the Cookie header.
// -b, --cookie <data|filename>
func cookieCmd(cookieCommandString string) (string, error) {

	fmt.Println("cookie")

	if _, err := os.Stat(cookieCommandString); err == nil {
		// path/to/whatever exists

	} else if os.IsNotExist(err) {
		// path/to/whatever does *not* exist

	} else {
		// Schrodinger: file may or may not exist. See err for details.

		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence

	}

	return "", nil
}

// cookieJarCMD Specify to which file you want curl to write all cookies after a completed operation.
// -c, --cookie-jar <filename>
func cookieJarCmd(cookieJarDirectory string) (string, error) {

	return "", nil
}

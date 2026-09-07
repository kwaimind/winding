// Command winding bumps every tracked repo to the latest Yarn and installs.
package main

import (
	"os"

	"github.com/kwaimind/winding/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1) // cobra already printed the error
	}
}

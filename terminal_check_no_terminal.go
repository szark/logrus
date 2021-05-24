// +build js nacl plan9

package logrus2

import (
	"io"
)

func checkIfTerminal(w io.Writer) bool {
	return false
}

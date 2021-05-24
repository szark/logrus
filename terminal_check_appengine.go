// +build appengine

package logrus2

import (
	"io"
)

func checkIfTerminal(w io.Writer) bool {
	return true
}

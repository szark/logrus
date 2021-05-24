package logrus

import (
	"runtime"
	"strings"
)

// oGetCaller returns the filename and the line info of a function
// further down in the call stack.  Passing 0 in as callDepth would
// return info on the function calling oGetCallerIgnoringLog, 1 the
// parent function, and so on.  Any suffixes passed to oGetCaller are
// path fragments like "/pkg/log/log.go", and functions in the call
// stack from that file are ignored.
func oGetCaller(callDepth int, suffixesToIgnore ...string) (file string, line int) {
	// bump by 1 to ignore the oGetCaller (this) stackframe
	callDepth++
outer:
	for {
		var ok bool
		_, file, line, ok = runtime.Caller(callDepth)
		if !ok {
			file = "???"
			line = 0
			break
		}

		for _, s := range suffixesToIgnore {
			if strings.HasSuffix(file, s) {
				callDepth++
				continue outer
			}
		}
		break
	}
	return
}

func oGetCallerIgnoringLogMulti(callDepth int) (string, int) {
	// the +1 is to ignore this (oGetCallerIgnoringLogMulti) frame
	return oGetCaller(callDepth+1, "/pkg/log/log.go", "/pkg/io/multi.go")
}

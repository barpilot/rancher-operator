// Package klogr implements github.com/go-logr/logr.Logger in terms of
// k8s.io/klog.
package klogr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"sort"

	"github.com/go-logr/logr"
	"k8s.io/klog"
)

// New returns a logr.Logger which is implemented by klog.
func New() logr.Logger {
	return klogger{
		level:  0,
		prefix: "",
		values: nil,
	}
}

type klogger struct {
	level  int
	prefix string
	values []interface{}
}

func (l klogger) clone() klogger {
	return klogger{
		level:  l.level,
		prefix: l.prefix,
		values: copySlice(l.values),
	}
}

func copySlice(in []interface{}) []interface{} {
	out := make([]interface{}, len(in))
	copy(out, in)
	return out
}

// Magic string for intermediate frames that we should ignore.
const autogeneratedFrameName = "<autogenerated>"

// Discover how many frames we need to climb to find the caller. This approach
// was suggested by Ian Lance Taylor of the Go team, so it *should* be safe
// enough (famous last words).
func framesToCaller() int {
	// 1 is the immediate caller.  3 should be too many.
	for i := 1; i < 3; i++ {
		_, file, _, _ := runtime.Caller(i + 1) // +1 for this function's frame
		if file != autogeneratedFrameName {
			return i
		}
	}
	return 1 // something went wrong, this is safe
}

func flatten(kvList ...interface{}) string {
	keys := make([]string, 0, len(kvList))
	vals := make(map[string]interface{}, len(kvList))
	for i := 0; i < len(kvList); i += 2 {
		k, ok := kvList[i].(string)
		if !ok {
			panic(fmt.Sprintf("key is not a string: %s", pretty(kvList[i])))
		}
		var v interface{}
		if i+1 < len(kvList) {
			v = kvList[i+1]
		}
		keys = append(keys, k)
		vals[k] = v
	}
	sort.Strings(keys)
	buf := bytes.Buffer{}
	for i, k := range keys {
		v := vals[k]
		if i > 0 {
			buf.WriteRune(' ')
		}
		buf.WriteString(pretty(k))
		buf.WriteString("=")
		buf.WriteString(pretty(v))
	}
	return buf.String()
}

func pretty(value interface{}) string {
	jb, _ := json.Marshal(value)
	return string(jb)
}

func (l klogger) Info(msg string, kvList ...interface{}) {
	if l.Enabled() {
		lvlStr := flatten("level", l.level)
		msgStr := flatten("msg", msg)
		fixedStr := flatten(l.values...)
		userStr := flatten(kvList...)
		klog.InfoDepth(framesToCaller(), l.prefix, " ", lvlStr, " ", msgStr, " ", fixedStr, " ", userStr)
	}
}

func (l klogger) Enabled() bool {
	return bool(klog.V(klog.Level(l.level)))
}

func (l klogger) Error(err error, msg string, kvList ...interface{}) {
	msgStr := flatten("msg", msg)
	var loggableErr interface{}
	if err != nil {
		loggableErr = err.Error()
	}
	errStr := flatten("error", loggableErr)
	fixedStr := flatten(l.values...)
	userStr := flatten(kvList...)
	klog.ErrorDepth(framesToCaller(), l.prefix, " ", msgStr, " ", errStr, " ", fixedStr, " ", userStr)
}

func (l klogger) V(level int) logr.InfoLogger {
	new := l.clone()
	new.level = level
	return new
}

// WithName returns a new logr.Logger with the specified name appended.  klogr
// uses '/' characters to separate name elements.  Callers should not pass '/'
// in the provided name string, but this library does not actually enforce that.
func (l klogger) WithName(name string) logr.Logger {
	new := l.clone()
	if len(l.prefix) > 0 {
		new.prefix = l.prefix + "/"
	}
	new.prefix += name
	return new
}

func (l klogger) WithValues(kvList ...interface{}) logr.Logger {
	new := l.clone()
	new.values = append(new.values, kvList...)
	return new
}

var _ logr.Logger = klogger{}
var _ logr.InfoLogger = klogger{}

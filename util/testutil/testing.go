// The MIT License (MIT)

// Copyright (c) 2014 Ben Johnson

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package testutil

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/goleak"
)

// This package is imported by non-test code and therefore cannot import the
// testing package, which has side effects such as adding flags. Hence we use an
// interface to testing.{T,B}.
type TB interface {
	Helper()
	Fatalf(string, ...interface{})
}

// Assert fails the test if the condition is false.
func Assert(tb TB, condition bool, format string, a ...interface{}) {
	tb.Helper()
	if !condition {
		tb.Fatalf("\033[31m"+format+"\033[39m\n", a...)
	}
}

// Ok fails the test if an err is not nil.
func Ok(tb TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("\033[31munexpected error: %v\033[39m\n", err)
	}
}

// NotOk fails the test if an err is nil.
func NotOk(tb TB, err error, a ...interface{}) {
	tb.Helper()
	if err == nil {
		if len(a) != 0 {
			format := a[0].(string)
			tb.Fatalf("\033[31m"+format+": expected error, got none\033[39m", a[1:]...)
		}
		tb.Fatalf("\033[31mexpected error, got none\033[39m")
	}
}

// Equals fails the test if exp is not equal to act.
func Equals(tb TB, exp, act interface{}, msgAndArgs ...interface{}) {
	tb.Helper()
	if diff := cmp.Diff(exp, act,
		cmp.Exporter(func(reflect.Type) bool { return true }), // Compare unexported fields.
	); diff != "" {
		tb.Fatalf("\033[31m%s\n\nexp: %#v\n\ngot: %#v\033[39m\n\ndiff:\n\n%s\n", formatMessage(msgAndArgs), exp, act, diff)
	}
}

// ErrorEqual compares Go errors for equality.
func ErrorEqual(tb TB, left, right error, msgAndArgs ...interface{}) {
	tb.Helper()
	if left == right {
		return
	}

	if left != nil && right != nil {
		Equals(tb, left.Error(), right.Error(), msgAndArgs...)
		return
	}

	tb.Fatalf("\033[31m%s\n\nexp: %#v\n\ngot: %#v\033[39m\n", formatMessage(msgAndArgs), left, right)
}

func formatMessage(msgAndArgs []interface{}) string {
	if len(msgAndArgs) == 0 {
		return ""
	}

	if msg, ok := msgAndArgs[0].(string); ok {
		return fmt.Sprintf("\n\nmsg: "+msg, msgAndArgs[1:]...)
	}
	return ""
}

// TolerantVerifyLeak verifies go leaks but excludes the go routines that are
// launched as side effects of some of our dependencies.
func TolerantVerifyLeak(m *testing.M) {
	goleak.VerifyTestMain(m,
		// https://github.com/census-instrumentation/opencensus-go/blob/d7677d6af5953e0506ac4c08f349c62b917a443a/stats/view/worker.go#L34
		goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start"),
		// https://github.com/kubernetes/klog/blob/c85d02d1c76a9ebafa81eb6d35c980734f2c4727/klog.go#L417
		goleak.IgnoreTopFunction("k8s.io/klog/v2.(*loggingT).flushDaemon"),
	)
}

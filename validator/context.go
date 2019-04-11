// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package validator

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// ExecContext contains methods need for checkF and testF.
type ExecContext interface {
	Log(format string, args ...interface{})

	Fail(msg ...string)
	Assert(b bool, msg ...string)
	AssertNil(x interface{}, msg ...string)
	AssertNotNil(x interface{}, msg ...string)
	AssertEQ(x, y interface{}, msg ...string)
	AssertNE(x, y interface{}, msg ...string)
	AssertDeepEQ(x, y interface{}, msg ...string)
	AddCallerDepth(n int)
}

// Recorder records a execute history of checker or test.
type Recorder struct {
	Description string   `json:"description,omitempty"`
	Logs        []string `json:"logs,omitempty"`
	Success     bool     `json:"success,omitempty"`
}

func newRecorder(description string) *Recorder {
	return &Recorder{Description: description}
}

// Log records some message in the Recorder.
func (r *Recorder) Log(format string, args ...interface{}) {
	r.log(LogFileLine, format, args...)
}

func (r *Recorder) log(logCaller bool, format string, args ...interface{}) {
	var builder strings.Builder
	builder.WriteString(time.Now().Format(LogTimeFormat))
	if logCaller {
		builder.WriteByte(' ')
		builder.WriteString(caller(3))
	}
	builder.WriteByte(' ')
	builder.WriteString(fmt.Sprintf(format, args...))
	r.Logs = append(r.Logs, builder.String())
}

// copy returns a copy that is safe to append logs.
func (r *Recorder) copy() *Recorder {
	return &Recorder{
		Description: r.Description,
		Logs:        r.Logs[:len(r.Logs):len(r.Logs)],
		Success:     r.Success,
	}
}

type asserter struct {
	callerDepth int
}

func (a *asserter) Fail(msg ...string) {
	a.panicNow("test is marked failed", msg...)
}

func (a *asserter) Assert(b bool, msg ...string) {
	if !b {
		a.panicNow("assertion failed", msg...)
	}
}

func (a *asserter) AssertNil(x interface{}, msg ...string) {
	if x != nil {
		a.panicNow(fmt.Sprintf("expect nil, got %v", x), msg...)
	}
}

func (a *asserter) AssertNotNil(x interface{}, msg ...string) {
	if x == nil {
		a.panicNow(fmt.Sprintf("expect not nil, got: %v", x), msg...)
	}
}

func (a *asserter) AssertEQ(x, y interface{}, msg ...string) {
	if x != y {
		a.panicNow(fmt.Sprintf("expect equal, got %v and %v", x, y), msg...)
	}
}

func (a *asserter) AssertNE(x, y interface{}, msg ...string) {
	if x == y {
		a.panicNow(fmt.Sprintf("expect not equal, got %v and %v", x, y), msg...)
	}
}

func (a *asserter) AssertDeepEQ(x, y interface{}, msg ...string) {
	if !reflect.DeepEqual(x, y) {
		a.panicNow(fmt.Sprintf("expect equal, got %v and %v", x, y), msg...)
	}
}

func (a *asserter) AddCallerDepth(n int) {
	a.callerDepth += n
}

func (a *asserter) panicNow(msg string, userMessage ...string) {
	var position string
	if LogFileLine {
		position = caller(3+a.callerDepth) + " "
	}
	if len(userMessage) > 0 {
		panic(position + strings.Join(userMessage, ","))
	}
	panic(position + msg)
}

func caller(n int) string {
	_, f, l, ok := runtime.Caller(n)
	if ok {
		return fmt.Sprintf("%s:%v", filepath.Base(f), l)
	}
	return ""
}

type execContext struct {
	*Recorder
	*asserter
}

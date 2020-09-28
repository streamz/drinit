// +build linux
/*
Copyright © 2020 streamz <bytecodenerd@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package log

import (
	"fmt"
gol "log"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
)

// Level - severity level
type Level int

func (l Level) String() string {
	return lmap[l]
}

// levels.
const (
	TraceL Level = iota
	InfoL 
	WarnL
	ErrorL
	PanicL
)

const (
	_empty = ""
	_newln = "\n"
)

const (
	flags = gol.Ldate | gol.Lmicroseconds
)

var lmap = map[Level]string {
	TraceL: "TRACE",
	InfoL: 	"INFO",
	WarnL:	"WARN",
	ErrorL:	"ERROR",
	PanicL: "PANIC",
}

type callinfo struct {
	pkg string
    fil string
    fun string
    lne string
}
// Log object
type Log struct {
	lock 	sync.Mutex
	level 	Level
	stdout 	*gol.Logger
	stderr 	*gol.Logger
}

var logmap = make(map[string]*Log)
var mutex = sync.Mutex{}

// Logger - create a new logger. 
// The package name is used and the log is cached
func Logger() *Log {
	_, name, _ := info(2)
	log, ok := logmap[name]; if !ok {
		mutex.Lock()
		log = new()
		logmap[name] = log
		mutex.Unlock()
	}
	return log
}

func new() *Log {
	return &Log{
		level: InfoL,
		stdout: gol.New(os.Stdout, _empty, flags),
		stderr: gol.New(os.Stderr, _empty, flags),
	}
}

// Level set the loglevel
func (l *Log) Level(level Level) {
	l._level(level)
}

// Tracef logs with the TraceL loglevel.
func (l *Log) Tracef(fs string, v ...interface{}) {
	l.out(TraceL, 0, fmt.Sprintf(fs, v...))
}

// Trace logs with the TraceL loglevel.
func (l *Log) Trace(v ...interface{}) {
	l.out(TraceL, 0, fmt.Sprint(v...))
}

// IsTrace -
func (l *Log) IsTrace() bool {
	return l.level == TraceL
}

// Infof logs with the InfoL loglevel.
func (l *Log) Infof(fs string, v ...interface{}) {
	l.out(InfoL, 0, fmt.Sprintf(fs, v...))
}

// Info logs with the InfoL loglevel.
func (l *Log) Info(v ...interface{}) {
	l.out(InfoL, 0, fmt.Sprint(v...))
}

// IsInfo -
func (l *Log) IsInfo() bool {
	return l.level == InfoL
}

// Warnf logs with the WarnL loglevel.
func (l *Log) Warnf(fs string, v ...interface{}) {
	l.out(WarnL, 0, fmt.Sprintf(fs, v...))
}

// Warn logs with the WarnL loglevel.
func (l *Log) Warn(v ...interface{}) {
	l.out(WarnL, 0, fmt.Sprint(v...))
}

// Errorf logs with the ErrorL loglevel.
func (l *Log) Errorf(fs string, v ...interface{}) {
	l.out(ErrorL, 0, fmt.Sprintf(fs, v...))
}

// Error logs with the ErrorL loglevel.
func (l *Log) Error(v ...interface{}) {
	l.out(ErrorL, 0, fmt.Sprint(v...))
}

// Panicf logs with the Panic loglevel.
func (l *Log) Panicf(fs string, v ...interface{}) {
	l.out(PanicL, 0, fmt.Sprintf(fs, v...))
	os.Exit(1)
}

// Panic logs with the Panic loglevel.
func (l *Log) Panic(v ...interface{}) {
	l.out(PanicL, 0, fmt.Sprint(v...))
	os.Exit(1)
}

func (l *Log) _level(level Level) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.level = level 
}

func (l *Log) out(level Level, depth int, str string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	msg := fmt.Sprintf("%s: %s - %s", level.String(), infostr(), str)
	if l.level <= level {
		switch level {
		case TraceL, InfoL:
			l.stdout.Output(3+depth, msg + _newln)
		case WarnL, ErrorL, PanicL:
			l.stderr.Output(3+depth, msg + _newln)
		}
	}
}

func info(depth int) (string, string, int) {
    pc, file, line, _ := runtime.Caller(depth)
    _, fname := path.Split(file)
    parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
    pl := len(parts)
    pkg := ""

    if parts[pl-2][0] == '(' {
        pkg = strings.Join(parts[0:pl-2], ".")
    } else {
        pkg = strings.Join(parts[0:pl-1], ".")
    }

	return fname, pkg, line
}

func infostr() string {
	f, p, l := info(4)
	return fmt.Sprintf("%s %s line: %d", f, p, l)
}

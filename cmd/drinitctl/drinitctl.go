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

package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/streamz/drinit/cli"
	"github.com/streamz/drinit/ipc"
	"github.com/streamz/drinit/log"
	"github.com/streamz/drinit/sig"
)

func main() {
	c := newcli()
	l := log.Logger()
	m := map[int]string{
		_cycle: ipc.Cycle,
		_down: ipc.Down,
		_up: ipc.Up,
	}

	l.Level(c.level)
	l.Infof("drinitctl %s", c.String())

	switch c.ctlmode {
	case _proc:
		cmd := m[c.command]
		l.Tracef("cmd: %s run %s", cmd, strings.Join(c.run, " "))
		msg := ipc.Msg{
			Name: cmd,
			Args: c.run,
		}
		ipc.Send(c.namedpipe, msg)
	case _signal:
		l.Tracef("sending signal %v to service", c.signal)
		msg := ipc.Msg{
			Name: ipc.Signal,
			Args: []string{sig.SignalToName(c.signal)},
		}
		ipc.Send(c.namedpipe, msg)
	}
}

// helpers
const verbosemsg = "verbose logging"
const helpemsg = "displays help usage"
const fdmsg = "the ipc named pipe"
const signalmsg = "send a signal to the supervised process"
const commandmsg = "1 - CYCLE, 2 - UP or 3 - DOWN the supervised service"
const runmsg = "the command to run before DOWN, after UP service command"
const usage = "/drinitctl -c2 -r echo stopping"

const (
	// cycle the service
	_cycle = iota + 1
	// start the service
	_up
	// sStop the service
	_down
)

type mode int

func (m mode) String() string {
	switch m {
	case _proc:
		return "PROC"
	case _signal:
		return "SIGNAL"
	}
	return "INVALID"
}

const (
	_invalid mode = iota
	// change process state (_cycle, _up. _down)
	_proc
	// signal child
	_signal
)

type clictx struct {
	level		log.Level
	namedpipe	string
	command		int
	signal		syscall.Signal
	ctlmode		mode
	run			[]string
}

func (c *clictx) String() string {
	return fmt.Sprintf(
		"level: %s, namedpipe: %s, command: %d, signal: %s, mode: %s, run: %v",
		c.level.String(), c.namedpipe, c.command, c.signal.String(), c.ctlmode.String(), c.run)
}

func newcli() *clictx {
	cmd := cli.New("drinitctl")
	help := cmd.Bool("help", "h", false, helpemsg)
	pipe := cmd.String("fd", "f", "/tmp/drinit.pipe", fdmsg)
	signal := cmd.String("signal", "s", "", signalmsg)
	command := cmd.Int("command", "c", 0, commandmsg)
	verbose := cmd.Bool("verbose", "v", false, verbosemsg)
	run := cmd.String("run", "r", "", runmsg)
	exit := func() {
		cmd.Usage(usage)
		os.Exit(0)
	}

	e := cmd.Parse()
	if e != nil {
		println("error %s, %s", e.Error(), strings.Join(os.Args, " "))
		exit()
	}

	if *help {
		cmd.Usage(usage)
		exit()
	}

	level := log.ErrorL
	if *verbose {
		level = log.TraceL
	}

	ctx := &clictx{
		level: level,
		namedpipe: *pipe,
		ctlmode: _invalid,
		run: []string{},
	}

	// order of preference if cmd has multiple options
	if len(*signal) > 1 {
		ctx.ctlmode = _signal
	}
	if *command >= _cycle && *command <= _down {
		ctx.ctlmode = _proc
		ctx.command = *command
	}

	switch ctx.ctlmode {
	case _proc:
		if len(*run) > 0 {
			ctx.run = strings.Split(strings.Trim(*run, " "), " ")
		}
	case _signal:
		s, e := sig.ToSignal(*signal)
		if e != nil {
			exit()
		}
		ctx.signal = s.(syscall.Signal)
	case _invalid:
		exit()
	}
	return ctx
}

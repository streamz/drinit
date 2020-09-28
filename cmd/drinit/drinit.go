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
	"os"
	"os/user"
	"strconv"
	"syscall"

	"github.com/streamz/drinit/exe"
	"github.com/streamz/drinit/ini"
	"github.com/streamz/drinit/log"
	"github.com/streamz/drinit/sig"
)

func main() {
	l := log.Logger()
	c := ini.NewCli()
	u, e := user.Current()
	if e != nil {
		l.Panic(e.Error())
	}

	l.Infof("drinit %s", c.String())

	h := func(s os.Signal) error {
		args := c.TrapArgs
		sz := len(args)

		l.Tracef("received signal: %s, trap args: %d", s.String(), sz)
		if sz == 0 {
			return nil
		}

		exc := exe.New(u)
		switch sz {
		case 1: // runs a shell script and passed signal is $1
			l.Tracef("invoking cmd: %s with signal: %s", args[0], s.String())
			return exc.Run(args[0], strconv.Itoa(sig.SignalToNumber(s.(syscall.Signal)))).Error
		default:
			l.Tracef("invoking cmd: %s with args: %+v", args[0], args[1:])
			return exc.Run(args[0], args[1:]...).Error
		}
	}

	o := &ini.InitOpts{
		Traps:  c.Traps,
		Signf:  h,
		Osuser: u,
	}

	i := ini.New(c.Supervise, c.Pipe, o)
	i.Start()
}

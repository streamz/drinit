// +build linux
// Copyright (c) 2015 ramr
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

package exe

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/streamz/drinit/log"
)

// Reaper - a zombie process reaper
type Reaper struct {
	log *log.Log
	one sync.Once
}

// NewReaper - Constructor for a zombie process reaper
func NewReaper() *Reaper {
	return &Reaper{
		log: log.Logger(),
		one: sync.Once{},
	}
}

func (r *Reaper) sigchldH(notifier chan os.Signal) {
	sigq := make(chan os.Signal, 1)
	signal.Notify(sigq, syscall.SIGCHLD)

	for {
		sig := <-sigq
		select {
		case notifier <- sig:
		default:
		}
	}
}

// Start - Start the reaper goroutine
func (r *Reaper) Start() {
	r.one.Do(func() {
		go r.reap()
	})
}

func (r *Reaper) reap() {
	notifier := make(chan os.Signal, 1)
	go r.sigchldH(notifier)

	pid := -1
	opt := 0

	for {
		sig := <-notifier
		r.log.Tracef("received signal %s", sig)

		for {
			var wstatus syscall.WaitStatus
			var rusage syscall.Rusage

			pid, err := syscall.Wait4(pid, &wstatus, opt, &rusage)

			for err == syscall.EINTR {
				pid, err = syscall.Wait4(pid, &wstatus, opt, &rusage)
			}

			if err == syscall.ECHILD {
				break
			}

			r.log.Tracef("reap: pid=%d, wstatus=%+v, rusage=%v\n", pid, wstatus, rusage)
		}
	}
}

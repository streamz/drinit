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

package sig

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/streamz/drinit/log"
	"github.com/streamz/drinit/util"
)

// Signalf - type for a signal handler
type Signalf = func(os.Signal) error

// SignalOpts -
type SignalOpts struct {
	Trapf Signalf
	Fwrdf Signalf
	Traps []os.Signal
}

// Signalh - internal signal handle state
type Signalh struct {
	logr *log.Log
	strt sync.Once
	init util.AtomicBool
	sigs map[os.Signal]struct{}
	sigf Signalf
	fwdf Signalf
	done context.Context
	Stop context.CancelFunc
}

// New - Create a new signal handler
func New(opts SignalOpts) *Signalh {
	noop := func(os.Signal) error { return nil }
	traps := opts.Traps
	trapf := opts.Trapf
	fwrdf := opts.Fwrdf
	logr := log.Logger()

	if traps == nil {
		traps = []os.Signal{}
	}

	if trapf == nil {
		trapf = noop
		if len(traps) > 0 {
			logr.Warn("signal handler is nil, ignoring traps")
			traps = []os.Signal{}
		}
	}

	if fwrdf == nil {
		logr.Warn("forward handler is nil, signals will not be forwarded to children")
		fwrdf = noop
	}

	signals := make(map[os.Signal]struct{})
	for _, s := range traps {
		signals[s] = struct{}{}
	}

	ctx, stop := context.WithCancel(context.Background())
	return &Signalh{
		logr: logr,
		strt: sync.Once{},
		init: util.AtomicBool{},
		sigs: signals,
		sigf: trapf,
		fwdf: fwrdf,
		done: ctx,
		Stop: stop,
	}
}

// Start - start the signal handler
func (h *Signalh) Start() error {
	init := h.init.Get()

	h.strt.Do(func() {
		h.init.Set()
		go h.handle()
	})

	if init {
		return fmt.Errorf("handler is already running")
	}
	return nil
}

func (h *Signalh) handle() {
	c := make(chan os.Signal, 1)
	signal.Notify(c)
	// Ignore SIGURG as it is used for goroutine preemption
	signal.Reset(syscall.SIGURG)

	for {
		select {
		case sig := <-c:
			if h.logr.IsTrace() {
				h.logr.Tracef("received signal %v", sig)
			}
			_, trap := h.sigs[sig]
			if trap {
				h.sigf(sig)
			} else {
				h.fwdf(sig)
			}
		case <-h.done.Done():
			h.logr.Trace("handler exiting")
			return
		}
	}
}

var signal2name = map[syscall.Signal]string{
	syscall.SIGHUP:    "SIGHUP",
	syscall.SIGINT:    "SIGINT",
	syscall.SIGQUIT:   "SIGQUIT",
	syscall.SIGILL:    "SIGILL",
	syscall.SIGTRAP:   "SIGTRAP",
	syscall.SIGABRT:   "SIGABRT",
	syscall.SIGBUS:    "SIGBUS",
	syscall.SIGFPE:    "SIGFPE",
	syscall.SIGKILL:   "SIGKILL",
	syscall.SIGUSR1:   "SIGUSR1",
	syscall.SIGSEGV:   "SIGSEGV",
	syscall.SIGUSR2:   "SIGUSR2",
	syscall.SIGPIPE:   "SIGPIPE",
	syscall.SIGALRM:   "SIGALRM",
	syscall.SIGTERM:   "SIGTERM",
	syscall.SIGCHLD:   "SIGCHLD",
	syscall.SIGCONT:   "SIGCONT",
	syscall.SIGSTOP:   "SIGSTOP",
	syscall.SIGTSTP:   "SIGTSTP",
	syscall.SIGTTIN:   "SIGTTIN",
	syscall.SIGTTOU:   "SIGTTOU",
	syscall.SIGURG:    "SIGURG",
	syscall.SIGXCPU:   "SIGXCPU",
	syscall.SIGXFSZ:   "SIGXFSZ",
	syscall.SIGVTALRM: "SIGVTALRM",
	syscall.SIGPROF:   "SIGPROF",
	syscall.SIGWINCH:  "SIGWINCH",
	syscall.SIGSYS:    "SIGSYS",
}

var name2signal map[string]os.Signal

// ToSignal string to Signal
func ToSignal(name string) (os.Signal, error) {
	if sig, ok := name2signal[name]; ok {
		return sig, nil
	}
	return nil, fmt.Errorf("invalid signal name: %s", name)
}

// SignalToName -
func SignalToName(s os.Signal) string {
	return signal2name[s.(syscall.Signal)]
}

// SignalToNumber - only works for *nix
func SignalToNumber(s os.Signal) int {
	return int(s.(syscall.Signal))
}

func init() {
	name2signal = make(map[string]os.Signal)
	for v, k := range signal2name {
		name2signal[k] = v
	}
}

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

package ini

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"sync"
	"syscall"
	"time"

	"github.com/streamz/drinit/exe"
	"github.com/streamz/drinit/ipc"
	"github.com/streamz/drinit/log"
	"github.com/streamz/drinit/sig"
)

// InitOpts -
type InitOpts struct {
	Traps []string
	Signf sig.Signalf
	Delay time.Duration
	Osusr *user.User
}

// Init - The supervisor proces handle
type Init struct {
	log *log.Log
	ctx context.Context
	can context.CancelFunc
	lok *sync.RWMutex
	ipc *ipc.Pipe
	sig *sig.Signalh
	rpr *exe.Reaper
	exc *exe.Exe
	syn sync.Once
	dly time.Duration
	cmd []string
}

type muxer map[string]func(*Init, []string)

// New - Constructor
func New(cmd []string, fd string, opts *InitOpts) *Init {
	cl := make([]string, len(cmd))
	copy(cl, cmd)

	ctx, can := context.WithCancel(context.Background())

	i := &Init{
		log: log.Logger(),
		ctx: ctx,
		can: can,
		lok: &sync.RWMutex{},
		rpr: exe.NewReaper(),
		exc: exe.New(opts.Osusr),
		syn: sync.Once{},
		dly: opts.Delay,
		cmd: cl,
	}

	i.sig = signalhandler(i, opts)

	var err error
	i.ipc, err = ipc.New(fd)
	if err != nil {
		i.log.Panic(err.Error())
	}

	return i
}

// Close -
func Close(i *Init) {
	i.can()
}

// Start - Starts the supervised program
func (i *Init) Start() {
	i.syn.Do(func() {
		i.rpr.Start()
		if e := i.sig.Start(); e != nil {
			i.log.Panic(e.Error())
		}
		go func(init *Init) {
			started, _ := init.exc.Start(init.cmd[0], init.cmd[1:]...)
			ok := <-started
			if !ok {
				init.log.Panicf("failed to start program, +%v", init.exc.Info())
			}
		}(i)
		i.service()
	})
}

func (i *Init) service() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Kill)

	// ipc listen
	recv := i.ipc.Open()
	mux := newmuxer()

	for {
		select {
		case msg := <-recv:
			i.log.Tracef("ipc received message %+v", msg)
			var e error
			fn, ok := mux[msg.Name]
			if !ok {
				e = fmt.Errorf("unknown cmd %s", msg.Name)
			}
			if e == nil {
				fn(i, msg.Args)
			} else {
				i.log.Error(e.Error())
			}
		case <-c:
		case <-i.ctx.Done():
			i.shutdown()
			return
		default:
		}
	}
}

func (i *Init) signal(sig os.Signal) error {
	i.lok.RLock()
	defer i.lok.RUnlock()

	inf := i.exc.Info()
	if inf.Pid == -1 {
		return errors.New("os: process already released")
	}

	if inf.Pid == 0 {
		return errors.New("os: process not initialized")
	}

	if inf.Finished.Get() {
		return errors.New("os: process already finished")
	}

	s, ok := sig.(syscall.Signal)
	if !ok {
		return errors.New("os: unsupported signal type")
	}

	if e := syscall.Kill(-inf.Pid, s); e != nil {
		if e == syscall.ESRCH {
			return errors.New("os: process already finished")
		}
		return e
	}
	return nil
}

// programpid - for testing
func (i *Init) programpid() int {
	i.lok.Lock()
	defer i.lok.Unlock()
	return i.exc.Info().Pid
}

func (i *Init) join() <-chan struct{} {
	return i.exc.Join()
}

func (i *Init) shutdown() {
	_ = stop(i)
	i.sig.Stop()
	i.ipc.Close()
}

func signalhandler(i *Init, opts *InitOpts) *sig.Signalh {
	ntraps := 0
	if opts.Traps != nil {
		ntraps = len(opts.Traps)
	}

	var traps []os.Signal
	if ntraps > 0 {
		for _, t := range opts.Traps {
			s, e := sig.ToSignal(t)
			if e != nil {
				i.log.Panic(e.Error())
			} else {
				traps = append(traps, s)
			}
		}
	}

	sopts := sig.SignalOpts{
		Trapf: opts.Signf,
		Fwrdf: func(signal os.Signal) error {
			if signal == syscall.SIGCHLD {
				return nil
			}
			return i.signal(signal)
		},
		Traps: traps,
	}
	return sig.New(sopts)
}

func start(i *Init) error {
	info := i.exc.Info()

	if info.StartT != 0 && !(info.Finished.Get() || info.Signaled.Get()) {
		return fmt.Errorf("start failed, current running pid is: %d", info.Pid)
	}

	i.lok.Lock()
	i.exc = i.exc.Copy()
	i.lok.Unlock()

	time.Sleep(i.dly)

	start, ctx := i.exc.Start(i.cmd[0], i.cmd[1:]...)
	ok := <-start

	if !ok {
		info = <-ctx
		return fmt.Errorf("+%v", info)
	}
	return nil
}

func stop(i *Init) error {
	info := i.exc.Info()
	if info.Finished.Get() || info.Signaled.Get() {
		return fmt.Errorf("stop failed, process is not running")
	}

	time.Sleep(i.dly)
	if err := i.exc.Terminate(); err != nil {
		return err
	}

	<-i.join()
	return nil
}

func restart(i *Init) error {
	i.lok.Lock()
	defer i.lok.Unlock()

	wait := i.exc.Join()
	err := i.exc.Terminate()
	if err == nil {
		// if the program has already terminated, we just launch a new one
		// otherwise, we wait until termination is complete
		<-wait
	}

	i.exc = i.exc.Copy()
	time.Sleep(i.dly)

	start, ctx := i.exc.Start(i.cmd[0], i.cmd[1:]...)
	ok := <-start

	if !ok {
		info := <-ctx
		return fmt.Errorf("+%v", info)
	}
	return nil
}

func sigp(i *Init, sig syscall.Signal) error {
	i.lok.Lock()
	defer i.lok.Unlock()

	info := i.exc.Info()
	if info.StartT == 0 || info.Finished.Get() || info.Signaled.Get() {
		return fmt.Errorf("signal failed, process is not running")
	}
	return syscall.Kill(-info.Pid, sig)
}

func runproc(args []string) *exe.Info {
	sz := len(args)
	if sz > 0 {
		switch sz {
		case 1:
			return exe.New(nil).Run(args[0])
		default:
			return exe.New(nil).Run(args[0], args[1:]...)
		}
	}
	return nil
}

func newmuxer() muxer {
	mux := make(muxer)
	mux[ipc.Signal] = func(i *Init, args []string) {
		s := ""
		if len(args) == 1 {
			s = args[0]
		}
		sign, e := sig.ToSignal(s)
		if e != nil {
			i.log.Error(e.Error())
		}
		if e = sigp(i, sign.(syscall.Signal)); e != nil {
			i.log.Error(e.Error())
		}
	}
	mux[ipc.Up] = func(i *Init, args []string) {
		if e := start(i); e != nil {
			i.log.Error(e.Error())
		}
		if len(args) > 0 {
			if info := runproc(args); info.Error != nil {
				i.log.Error(info.Error.Error())
			}
		}
	}
	mux[ipc.Down] = func(i *Init, args []string) {
		if len(args) > 0 {
			if info := runproc(args); info.Error != nil {
				i.log.Error(info.Error.Error())
			}
		}
		if e := stop(i); e != nil {
			i.log.Error(e.Error())
		}
	}
	mux[ipc.Cycle] = func(i *Init, args []string) {
		if e := restart(i); e != nil {
			i.log.Error(e.Error())
		}
	}
	return mux
}

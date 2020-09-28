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

package exe

import (
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/streamz/drinit/log"
	"github.com/streamz/drinit/util"
)

type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

// Info -
type Info struct {
	Pid			int
	Exit		int
	Error		error
	Finished 	util.AtomicBool
	Signaled 	util.AtomicBool
	RunT		time.Duration
	StartT 		int64
	EndT		int64
}

type status int

const (
	_uninitialized status = iota
	_exited
	_running
	_signaled
)

// Exe -
type Exe struct {
	log *log.Log
	lok *sync.Mutex
	usr *user.User
	ini *sync.Once
	sta status
	inf Info
	str time.Time
	ech chan Info
	sch chan bool
	syn chan struct{}
	ncp noCopy
}

// New -
func New(usr *user.User) *Exe {
	osuser := usr
	if osuser == nil {
		osuser, _ = user.Current()
	}
	return &Exe{
		log: log.Logger(),
		usr: osuser,
		lok: &sync.Mutex{},
		ini: &sync.Once{},
		inf: Info{Pid: 0, Exit: -1},
		sta: _uninitialized,
		ech: make(chan Info, 1),
		sch: make(chan bool, 1),
		syn: make(chan struct{}),
	}
}

// Start -
func (x *Exe) Start(name string, args ...string) (<-chan bool, <-chan Info) {
	x.ini.Do(func() {
		x.log.Tracef("running cmd: %s %s", name, strings.Join(args, " "))
		go x.runf(name, args...)
	})
	return x.sch, x.ech
}

// Run -
func (x *Exe) Run(name string, args ...string) *Info {
	x.log.Tracef("running cmd: %s %s", name, strings.Join(args, " "))
	_, complete := x.Start(name, args...)
	info := <-complete
	return &info
}

// Terminate -
func (x *Exe) Terminate() error {
	x.lok.Lock()
	defer x.lok.Unlock()

	if x.sta == _uninitialized || x.inf.Finished.Get() {
		return nil
	}

	x.sta = _signaled
	x.inf.Signaled.Set()
	return syscall.Kill(-x.inf.Pid, syscall.SIGTERM)
}

// Info -
func (x *Exe) Info() Info {
	x.lok.Lock()
	defer x.lok.Unlock()

	switch x.sta {
	case _running:
		x.inf.RunT = time.Now().Sub(x.str)
	case _exited:
		x.inf.Finished.Set()
	}
	return x.inf
}

// Copy -
func (x *Exe) Copy() *Exe {
	return New(x.usr)
}

// Join -
func (x *Exe) Join() <-chan struct{} {
	return x.syn
}

func (x *Exe) runf(name string, args ...string) {
	defer func() {
		x.ech <- x.Info()
		close(x.syn)
	}()

	cmd := x.newcmd(name, args...)
	now := time.Now()

	if e := cmd.Start(); e != nil {
		x.complete(&now, e)
		x.sch <- false
		return
	}

	x.init(&now, cmd)
	x.sch <- true
	err := cmd.Wait()
	x.complete(&now, err)
}

func (x *Exe) newcmd(name string, args ...string) *exec.Cmd {
	uid, _ := strconv.Atoi(x.usr.Uid)
	gid, _ := strconv.Atoi(x.usr.Gid)

	cred := &syscall.Credential{
		Uid: uint32(uid),
		Gid: uint32(gid),
		NoSetGroups: true,
	}

	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: cred,
		Setpgid: true,
	}

	cmd.Env = os.Environ()
	cmd.Dir = os.Getenv("PWD")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func (x *Exe) init(t *time.Time, cmd *exec.Cmd) {
	x.lok.Lock()
	defer x.lok.Unlock()

	x.inf.Pid = cmd.Process.Pid
	x.inf.StartT = t.UnixNano()
	x.sta = _running
}

func (x *Exe) complete(t *time.Time, err error) {
	code := 0
	if err != nil {
		code = exiterr(err)
	}
	x.endstate(t, code, err)
}

func (x *Exe) endstate(t *time.Time, code int, err error) {
	x.lok.Lock()
	defer x.lok.Unlock()

	x.inf.Error = err
	x.inf.Exit = code
	x.inf.StartT = t.UnixNano()
	x.inf.EndT = time.Now().UnixNano()
	if x.sta != _signaled {
		x.inf.Finished.Set()
		x.sta = _exited
	}
}

func exiterr(err error) int {
	if e, ok := err.(*exec.ExitError); ok {
		ws := e.Sys().(syscall.WaitStatus)
		if sig := ws.Signal(); sig > 0 {
			return int(sig)
		}
		return ws.ExitStatus()
	}
	return 0
}

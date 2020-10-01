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
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/streamz/drinit/ipc"
	"github.com/streamz/drinit/util"
	"github.com/stretchr/testify/assert"
)

var (
	_, b, _, _ = runtime.Caller(0)
	Testdata = strings.TrimSuffix(filepath.Dir(b), "/ini") + "/testdata/"
)

func TestStop(t *testing.T) {
	i := New(
		[]string{Testdata + "service.sh"},
		"/tmp/drinit-test-stop.pipe",
		&InitOpts{})

	completer := i.join()

	go i.Start()
	time.Sleep(time.Second)

	err := stop(i)
	assert.NoError(t, err)
	<-completer

	info := i.exc.Info()
	assert.False(t, info.Finished.Get(), "info should not be finished")
	assert.True(t, info.Signaled.Get(), "info should be Signaled")
	assert.Equal(
		t,
		15,
		info.Exit,
		"should exit with 15")
	Close(i)
}

func TestStopByPipe(t *testing.T) {
	f := "/tmp/drinit-test-stop-pipe.pipe"
	i := New(
		[]string{Testdata + "service.sh"},
		f,
		&InitOpts{})

	completer := i.join()

	go i.Start()
	time.Sleep(time.Second)

	ipc.Send(f, ipc.Msg{Name: ipc.Down})
	<-completer

	info := i.exc.Info()
	assert.False(t, info.Finished.Get(), "info should not be finished")
	assert.True(t, info.Signaled.Get(), "info should be Signaled")
	assert.Equal(
		t,
		15,
		info.Exit,
		"should exit with 15")
	Close(i)
}

func TestStart(t *testing.T) {
	i := New(
		[]string{Testdata + "service.sh"},
		"/tmp/drinit-test-start.pipe",
		&InitOpts{})

	joiner := i.join()

	go i.Start()
	time.Sleep(time.Second)

	err := stop(i)
	assert.NoError(t, err)
	<-joiner

	err = start(i)

	info := i.exc.Info()
	assert.NotEqual(t, 0, info.StartT)
	assert.False(t, info.Finished.Get(), "info should not be finished")
	assert.False(t, info.Signaled.Get(), "info should not be Signaled")

	Close(i)
}

func TestRestart(t *testing.T) {
	i := New(
		[]string{Testdata + "service.sh"},
		"/tmp/drinit-test-restart.pipe",
		&InitOpts{})

	joiner := i.join()

	go i.Start()
	time.Sleep(time.Second)

	info := i.exc.Info()
	oldpid := info.Pid

	err := restart(i)
	assert.NoError(t, err)
	<-joiner

	info = i.exc.Info()
	newpid := info.Pid

	assert.NotEqual(t, oldpid, newpid, "pids should not be equal")
	assert.NotEqual(t, 0, info.StartT, "startT should not be 0")
	assert.False(t, info.Finished.Get(), "info should not be finished")
	assert.False(t, info.Signaled.Get(), "info should not be Signaled")

	Close(i)
}

func TestSignal(t *testing.T) {
	w := &sync.WaitGroup{}
	b := &util.AtomicBool{}
	i := New(
		[]string{Testdata + "service.sh"},
		"/tmp/drinit-test-signal.pipe",
		&InitOpts{
			Traps: []string{"SIGUSR1"},
			Signf: func(sig os.Signal) error {
				switch sig {
				case syscall.SIGUSR1:
					b.Set()
					w.Done()
				}
				return nil
			},
		})

	w.Add(1)

	joiner := i.join()

	go i.Start()
	time.Sleep(time.Second)

	// send SIGUSR1 to the parent
	syscall.Kill(os.Getpid(), syscall.SIGUSR1)

	w.Wait()
	stop(i)
	<-joiner

	assert.True(t, b.Get(), "should be true")
	Close(i)
}

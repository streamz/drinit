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
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/streamz/drinit/exe"
	"github.com/stretchr/testify/assert"
)

var (
	_, b, _, _ = runtime.Caller(0)
	Testdata   = strings.TrimSuffix(filepath.Dir(b), "/sig") + "/testdata/"
)

func TestTrapSignal(t *testing.T) {
	u, _ := user.Current()
	exc := exe.New(u)
	opts := SignalOpts{
		Traps: []os.Signal{syscall.SIGUSR1},
		Trapf: func(sig os.Signal) error {
			switch sig {
			case syscall.SIGUSR1:
				if e := exc.Terminate(); e != nil {
					t.Error(e.Error())
				}
			}
			return nil
		},
		Fwrdf: nil,
	}

	h := New(opts)
	h.Start()

	time.Sleep(time.Second)

	s, ctx := exc.Start(Testdata + "service.sh")
	started := <-s
	assert.True(t, started)

	// send SIGUSR1 to the parent
	syscall.Kill(os.Getpid(), syscall.SIGUSR1)
	info := <-ctx
	h.Stop()

	assert.Error(t, info.Error)
	assert.False(t, info.Finished.Get(), "info should not be finished")
	assert.True(t, info.Signaled.Get(), "info should be Signaled")
	assert.Equal(
		t,
		15,
		info.Exit,
		"should exit with 15")
}

func TestForwardSignal(t *testing.T) {
	u, _ := user.Current()
	exc := exe.New(u)
	opts := SignalOpts{
		Traps: nil,
		Trapf: nil,
		Fwrdf: func(sig os.Signal) error {
			switch sig {
			case syscall.SIGUSR1:
				if e := exc.Terminate(); e != nil {
					t.Error(e.Error())
				}
			}
			return nil
		},
	}

	h := New(opts)
	h.Start()
	time.Sleep(time.Second)

	s, ctx := exc.Start(Testdata + "service.sh")
	started := <-s
	assert.True(t, started)

	// send SIGUSR1 to the parent
	syscall.Kill(os.Getpid(), syscall.SIGUSR1)
	info := <-ctx
	h.Stop()

	assert.Error(t, info.Error)
	assert.False(t, info.Finished.Get(), "info should not be finished")
	assert.True(t, info.Signaled.Get(), "info should be Signaled")
	assert.Equal(
		t,
		15,
		info.Exit,
		"should exit with 15")
}

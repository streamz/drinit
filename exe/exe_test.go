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
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	_, b, _, _ = runtime.Caller(0)
	Testdata   = strings.TrimSuffix(filepath.Dir(b), "/exe") + "/testdata/"
)

func TestRun(t *testing.T) {
	u, _ := user.Current()
	exc := New(u)
	info := exc.Run(Testdata + "process.sh")
	assert.NoError(t, info.Error)
}

func TestStart(t *testing.T) {
	u, _ := user.Current()
	exc := New(u)

	_, ctx := exc.Start(Testdata + "process.sh")
	info := <-ctx

	assert.NoError(t, info.Error)
	assert.True(t, info.Finished.Get(), "info should be finished")
	assert.False(t, info.Signaled.Get(), "info should not be Signaled")
	assert.Equal(
		t,
		0,
		info.Exit,
		"should exit with 0")
}

func TestTerminate(t *testing.T) {
	u, _ := user.Current()
	exc := New(u)

	started, ctx := exc.Start(Testdata + "service.sh")
	<-started

	time.Sleep(time.Second)

	// terminate the process
	err := exc.Terminate()
	assert.NoError(t, err)

	info := <-ctx
	assert.Error(t, info.Error)
	assert.False(t, info.Finished.Get(), "info should not be finished")
	assert.True(t, info.Signaled.Get(), "info should be Signaled")
	assert.Equal(
		t,
		15,
		info.Exit,
		"should exit with 15")
}

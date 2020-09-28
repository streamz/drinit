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

package ipc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var expect = []Msg { 
	{
		Name: "start",
		Args: []string{"a", "b", "c"},
	},
	{
		Name: "stop",
		Args: []string{"d", "e", "f"},
	},
}

func TestPipe(t *testing.T) {
	fd := "/tmp/test.pipe"
	pipe, err := New(fd)
	assert.NoError(t, err)
	defer pipe.Close()

	recv := pipe.Open()

	go func() {
		Send(fd, expect[0])
		Send(fd, expect[1])
	}()

	msg0 := <-recv
	msg1 := <-recv
	assert.Equal(t, expect[0].Name, msg0.Name, "message Name should be equal")
	assert.Equal(t, expect[0].Args, msg0.Args, "message Args should be equal")
	assert.NotEqual(t, expect[0].epoc, msg0.epoc, "message epoch should be equal")
	assert.Equal(t, expect[1].Name, msg1.Name, "message Name should be equal")
	assert.Equal(t, expect[1].Args, msg1.Args, "message Args should be equal")
	assert.NotEqual(t, expect[1].epoc, msg1.epoc, "message epoch should be equal")
}

func TestLoopPipe(t *testing.T) {
	fd := "/tmp/test-loop.pipe"
	pipe, err := New(fd)
	assert.NoError(t, err)
	defer pipe.Close()

	recv := pipe.Open()

	for i := 0; i < 10; i++ {
		Send(fd, expect[i%2])
	}

	var i int
	for msg := range recv {
		assert.Equal(t, expect[i%2].Name, msg.Name, "message Name should be equal")
		assert.Equal(t, expect[i%2].Args, msg.Args, "message Args should be equal")
		i++
		if i == 10 {
			break
		}
	}
}
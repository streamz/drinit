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

package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	cmd := []string{
		"-string", "test",
		"-strings", "a, b, c",
		"-int", "1",
		"-ints", "0, 1, 2",
		"-bool",
		"-duration", "3s",
		"--", "unknown",
	}

	run(t, cmd)

	cmd = []string{
		"-s", "test",
		"-a", "a, b, c",
		"-i", "1",
		"-n", "0, 1, 2",
		"-b",
		"-d", "3s",
		"--", "unknown",
	}

	run(t, cmd)

	cmd = []string{
		"-s", "test",
		"-strings", "a, b, c",
		"-i", "1",
		"-n", "0, 1, 2",
		"-bool",
		"-d", "3s",
		"--", "unknown",
	}

	run(t, cmd)
}

func run(t *testing.T, cmd []string) {
	cli := New("test")
	sptr := cli.String("string", "s", "", "a string")
	sslc := cli.StringSlice("strings", "a", "a string slice")
	iptr := cli.Int("int", "i", 0, "an int")
	islc := cli.IntSlice("ints", "n", "an int slice")
	bptr := cli.Bool("bool", "b", false, "a boolean")
	dptr := cli.Duration("duration", "d", time.Second, "a duration")

	if e := cli.ParseSlice(cmd); e != nil {
		t.Errorf("could not parse command %s", strings.Join(cmd, " "))
	}

	narg := cli.NArg()

	assert.True(t, cli.Parsed(), "values parsed")
	assert.Equal(t, "test", *sptr, "should be equal")
	assert.Equal(t, []string{"a", "b", "c"}, *sslc, "should be equal")
	assert.Equal(t, 1, *iptr, "should be equal")
	assert.Equal(t, []int{0, 1, 2}, *islc, "should be equal")
	assert.Equal(t, true, *bptr, "should be equal")
	assert.Equal(t, 3*time.Second, *dptr, "should be equal")
	assert.Equal(t, 1, narg, "should be equal")
}

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

package log

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestLevel(t *testing.T) {
    l := Logger()
    l.Level(TraceL)

    assert.True(t, l.IsTrace(), "should be true")
    assert.Equal(t, TraceL.String(), "TRACE", "should be equal")

    l = Logger()
    assert.True(t, l.IsTrace(), "should be true")

    l.Level(InfoL)
    assert.False(t, l.IsTrace(), "should be false")

    l = Logger()
    assert.True(t, l.IsInfo(), "should be true")
}
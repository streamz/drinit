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

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAtomicBool(t *testing.T) {
	b := AtomicBool{}
	b.Set()
	assert.True(t, b.Get(), "should be true")
	b.Clear()
	assert.False(t, b.Get(), "should be false")
	assert.False(t, b.Swap(true), "should be false")
	assert.True(t, b.Get(), "should be true")
}

func TestAtomicInt32(t *testing.T) {
	zero := int32(0)
	val := int32(7)
	i32 := AtomicInt32{}
	i32.Set(val)
	assert.Equal(t, i32.Get(), val, "should be equal")
	i32.Clear()
	assert.Equal(t, i32.Get(), zero, "should be equal")
	assert.Equal(t, i32.Swap(val), zero, "should be equal")
	assert.Equal(t, i32.Get(), val, "should be equal")
}

func TestAtomicInt(t *testing.T) {
	zero := 0
	val := 7
	i := AtomicInt{}
	i.Set(val)
	assert.Equal(t, i.Get(), val, "should be equal")
	i.Clear()
	assert.Equal(t, i.Get(), zero, "should be equal")
	assert.Equal(t, i.Swap(val), zero, "should be equal")
	assert.Equal(t, i.Get(), val, "should be equal")
}

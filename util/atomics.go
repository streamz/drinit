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

import "sync/atomic"

// AtomicBool - Race condition free primative wrapper
type AtomicBool struct {
	boolean int32
}

// Get - returns the underlying primative
func (b *AtomicBool) Get() bool {
	return (atomic.LoadInt32(&(b.boolean)) != 0)
}

// Set - sets the underlying primative
func (b *AtomicBool) Set() {
	atomic.StoreInt32(&(b.boolean), int32(1))
}

// Swap - exchanges the underlying primative
func (b *AtomicBool) Swap(value bool) bool {
	var i int32 = 0
	if value {
		i = 1
	}
	return atomic.SwapInt32(&(b.boolean), int32(i)) != 0
}

// Clear - resets the underlying primative to its unitialized default value
func (b *AtomicBool) Clear() {
	atomic.StoreInt32(&(b.boolean), int32(0))
}

// AtomicInt32 - Race condition free primative wrapper
type AtomicInt32 struct {
	value int32
}

// Get - returns the underlying primative
func (i *AtomicInt32) Get() int32 {
	return atomic.LoadInt32(&(i.value))
}

// Set - sets the underlying primative
func (i *AtomicInt32) Set(value int32) {
	atomic.StoreInt32(&(i.value), value)
}

// Incr - atomically increments underlying primative
func (i *AtomicInt32) Incr() int32 {
	return atomic.AddInt32(&(i.value), 1)
}

// Decr - atomically decrements underlying primative
func (i *AtomicInt32) Decr() int32 {
	return atomic.AddInt32(&(i.value), -1)
}

// Swap - exchanges the underlying primative
func (i *AtomicInt32) Swap(value int32) int32 {
	return atomic.SwapInt32(&(i.value), value)
}

// Clear - resets the underlying primative to its unitialized default value
func (i *AtomicInt32) Clear() {
	i.Set(0)
}

// AtomicInt - Race condition free primative wrapper
type AtomicInt struct {
	value int64
}

// Get - returns the underlying primative
func (i *AtomicInt) Get() int {
	return int(atomic.LoadInt64(&(i.value)))
}

// Set - sets the underlying primative
func (i *AtomicInt) Set(value int) {
	atomic.StoreInt64(&(i.value), int64(value))
}

// Incr - atomically increments underlying primative
func (i *AtomicInt) Incr() int {
	return int(atomic.AddInt64(&(i.value), 1))
}

// Decr - atomically decrements underlying primative
func (i *AtomicInt) Decr() int {
	return int(atomic.AddInt64(&(i.value), -1))
}

// Swap - exchanges the underlying primative
func (i *AtomicInt) Swap(value int) int {
	return int(atomic.SwapInt64(&(i.value), int64(value)))
}

// Clear - resets the underlying primative to its unitialized default value
func (i *AtomicInt) Clear() {
	i.Set(0)
}

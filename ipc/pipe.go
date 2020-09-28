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
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/streamz/drinit/log"
	"github.com/streamz/drinit/util"
)

// Msg - an IPC message
type Msg struct {
	epoc int64
	Name string
	Args []string
}

// Epoch -
func (m *Msg) Epoch() int64 {
	return m.epoc
}

func (m *Msg) String() string {
	return fmt.Sprintf("%d %s %s", m.epoc, m.Name, strings.Join(m.Args, " "))
}

// Pipe - a pipe
type Pipe struct {
	file *os.File
	once sync.Once
	clsd util.AtomicBool
	mchn chan Msg
}

var _log = log.Logger()

// New - Create a new half duplex ipc
func New(desc string) (*Pipe, error) {
	os.RemoveAll(desc)
	syscall.Mkfifo(desc, 0600)

	flags := os.O_RDWR|os.O_CREATE|os.O_APPEND|syscall.O_NONBLOCK
	file, err := os.OpenFile(desc, flags, os.ModeNamedPipe)
	if err != nil {
		return nil, err
	}
	
	return &Pipe {
		file: file,
		once: sync.Once{},
		clsd: util.AtomicBool{},
		mchn: make(chan Msg, 1),
	}, nil
}

// Open - open a pipe, returns a channel to receive messages on
func (p *Pipe) Open() <-chan Msg {
	p.once.Do(func() {
		go func(p *Pipe) {
			reader := bufio.NewReader(p.file)
			for {
				// will block when empty
				s, e := reader.ReadString('\n'); if e != nil {
					if p.clsd.Get() {
						return
					}
					_log.Error(e.Error())
				}
				s = strings.Trim(s, "\n")
				sa := strings.Split(strings.Trim(s, " "), " ")
				epoc, err := strconv.ParseInt(sa[0], 10, 64); if err != nil {
					epoc = time.Now().Unix()
				}
				msg := Msg{
					epoc: epoc,
					Name: sa[1],
					Args: sa[2:],
				}
				p.mchn <- msg
			}	
		}(p)
	})
	return p.mchn
}

// Close - close the pipe
func (p *Pipe) Close() {
	defer os.RemoveAll(p.file.Name())
	p.clsd.Set()
	close(p.mchn)
	p.file.Close()
}

// Send - sends a message to the desc (file)
func Send(desc string, msg Msg) error {
	flags := os.O_WRONLY|os.O_APPEND|syscall.O_NONBLOCK
	w, e := os.OpenFile(desc, flags, os.ModeNamedPipe); if e != nil {
		return e
	}
	defer w.Close()

	msg.epoc = time.Now().Unix()
	_, e = w.WriteString(msg.String()+"\n"); if e != nil {
		_log.Errorf("%+v", e)
		return e
	}
	return nil
}
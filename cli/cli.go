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
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type sslice struct {
	slice []string
}

func (s *sslice) Set(value string) error {
	res := strings.Split(value, ",")
	for _, r := range res {
		s.slice = append(s.slice, strings.TrimSpace(r))
	}
	return nil
}

func (s *sslice) String() string {
	return fmt.Sprintf("%v", *s)
}

type islice struct {
	slice []int
}

func (s *islice) Set(value string) error {
	res := strings.Split(value, ",")
	for _, r := range res {
		i, _ := strconv.Atoi(strings.TrimSpace(r))
		s.slice = append(s.slice, i)
	}
	return nil
}

func (s *islice) String() string {
	return fmt.Sprintf("%v", *s)
}

// Cli - a flags wrapper supporting slices
type Cli struct {
	flags *flag.FlagSet
}

// New - returns a new CLI
func New(cmd string) *Cli {
	return &Cli{flag.NewFlagSet(cmd, flag.ContinueOnError)}
}

// Parsed - returns true if flags have been parsed
func (c *Cli) Parsed() bool {
	return c.flags.Parsed()
}

// Parse - parse the command line
func (c *Cli) Parse() error {
	return c.flags.Parse(os.Args[1:])
}

// ParseSlice - parse a slice of commands
func (c *Cli) ParseSlice(slice []string) error {
	return c.flags.Parse(slice)
}

// StringSlice - creates a flag that is parsed into a string slice
func (c *Cli) StringSlice(name, shortname, usage string) *[]string {
	s := new(sslice)
	c.flags.Var(s, name, usage)
	if len(shortname) > 0 {
		c.flags.Var(s, shortname, usage)
	}
	return &s.slice
}

// String - creates a flag that is parsed into a string
func (c *Cli) String(name, shortname, value, usage string) *string {
	var str string
	c.flags.StringVar(&str, name, value, usage)
	if len(shortname) > 0 {
		c.flags.StringVar(&str, shortname, value, usage)
	}
	return &str
}

// IntSlice - creates a flag that is parsed into an int slice
func (c *Cli) IntSlice(name, shortname, usage string) *[]int {
	s := new(islice)
	c.flags.Var(s, name, usage)
	if len(shortname) > 0 {
		c.flags.Var(s, shortname, usage)
	}
	return &s.slice
}

// Int - creates a flag that is parsed into an int
func (c *Cli) Int(name, shortname string, value int, usage string) *int {
	var i int
	c.flags.IntVar(&i, name, value, usage)
	if len(shortname) > 0 {
		c.flags.IntVar(&i, shortname, value, usage)
	}
	return &i
}

// Bool - creates a flag that is parsed into a bool
func (c *Cli) Bool(name, shortname string, value bool, usage string) *bool {
	var b bool
	c.flags.BoolVar(&b, name, value, usage)
	if len(shortname) > 0 {
		c.flags.BoolVar(&b, shortname, value, usage)
	}
	return &b
}

// Duration - creates a flag that is parsed into a time.Duration
func (c *Cli) Duration(name, shortname string, value time.Duration, usage string) *time.Duration {
	var d time.Duration
	c.flags.DurationVar(&d, name, value, usage)
	if len(shortname) > 0 {
		c.flags.DurationVar(&d, shortname, value, usage)
	}
	return &d
}

// Usage - wrapped usage function
func (c *Cli) Usage(str string) {
	o := c.flags.Output()
	fmt.Fprintf(o, "Usage of %s:\n", c.flags.Name())
	fmt.Fprintf(o, "ex: %s", str)
	c.flags.PrintDefaults()
}

// Arg - wrapped Arg function
func (c *Cli) Arg(i int) string {
	return c.flags.Arg(i)
}

// NArg - wrapped NArg function
func (c *Cli) NArg() int {
	return c.flags.NArg()
}

// Args - wrapped Args function
func (c *Cli) Args() []string {
	return c.flags.Args()
}

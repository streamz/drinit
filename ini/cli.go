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
	"fmt"
    "os"
	"strings"

	"github.com/streamz/drinit/cli"
	"github.com/streamz/drinit/log"
)

const runmsg = "the script or command to run for a trap task. if the cmd has 0 args, the signal that triggered it will be in $1"
const verbosemsg = "verbose logging"
const helpemsg = "displays help usage"
const trapmsg = "the signals to trap"
const fdmsg =  "the named pipe"
const usage = "/drinit -- /program -and -args"

// CliContext -
type CliContext struct {
    Pipe    string
	Supervise, TrapArgs, Traps []string
}

func (c CliContext) String() string {
    return fmt.Sprintf(
        "pipe: %v, program: %v, traps: %v, run: %v", 
        c.Pipe, c.Supervise, c.Traps, c.TrapArgs)
}

// NewCli -
func NewCli() *CliContext {
	cmd := cli.New("drinit")
    help := cmd.Bool("help", "h", false, helpemsg)
    pipe := cmd.String("fd", "f", "/tmp/drinit.pipe", fdmsg)
    traprun := cmd.String("run", "r", "", runmsg)
    traps := cmd.StringSlice("traps", "t", trapmsg)
    verbose := cmd.Bool("verbose", "v", false, verbosemsg)
    
    logger := log.Logger()
    e := cmd.Parse(); if e != nil {
        logger.Errorf("%s, %s", e.Error(), strings.Join(os.Args, " "))
        cmd.Usage(usage)
        os.Exit(1)
	} 
    
    if *help {
        cmd.Usage(usage)
        os.Exit(0)
    }
    
    if *verbose {
        logger.Level(log.TraceL)
    }

    program := cmd.Args()
	if len(program) == 0 {
        logger.Error("program not defined")
        cmd.Usage(usage)
        os.Exit(1)
	} 
	
	return &CliContext {
        Pipe: *pipe,
		Supervise: program,
        TrapArgs: strings.Split(strings.Trim(*traprun, " "), " "),
        Traps: *traps,
	}
}

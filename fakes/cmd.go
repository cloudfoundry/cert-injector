package fakes

import (
	"fmt"
	"strings"
)

/* Represents all calls to a binary */
type Call struct {
	CallCount    int
	RecievedArgs []string
	ReturnVal    error
}

type RunCall struct {
	Calls map[string]*Call
}

func (r *RunCall) Returns(execBin string, toBeReturned error) {
	if _, ok := r.Calls[execBin]; !ok {
		r.Calls[execBin] = &Call{}
	}
	call := r.Calls[execBin]
	fmt.Println(toBeReturned)
	call.ReturnVal = toBeReturned
}

func (r *RunCall) CallCount(execBin string) int {
	call := r.Calls[execBin]
	return call.CallCount
}

func (r *RunCall) ReceivedArgs(execBin string) []string {
	call := r.Calls[execBin]
	return call.RecievedArgs
}

type Cmd struct {
	RunCalls *RunCall
}

func NewCmd() *Cmd {
	cmd := &Cmd{}
	cmd.RunCalls = &RunCall{}
	cmd.RunCalls.Calls = make(map[string]*Call)
	return cmd
}

func (c *Cmd) Run(executable string, args ...string) error {
	execPath := strings.Split(executable, "\\")
	execBin := execPath[len(execPath)-1]
	if _, ok := c.RunCalls.Calls[execBin]; !ok {
		c.RunCalls.Calls[execBin] = &Call{}
	}
	call := c.RunCalls.Calls[execBin]
	call.RecievedArgs = args
	call.CallCount++
	return call.ReturnVal
}

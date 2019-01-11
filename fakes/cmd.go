package fakes

import (
	"strings"
)

type RetVal struct {
	Stdout []byte
	Stderr []byte
	Err    error
}

/* Represents all calls to a binary */
type Call struct {
	CallCount    int
	RecievedArgs []string
	ReturnVal    *RetVal
}

type RunCall struct {
	Calls map[string]*Call
}

func (r *RunCall) initCall(execBin string) {
	if _, ok := r.Calls[execBin]; !ok {
		r.Calls[execBin] = &Call{}
		r.Calls[execBin].ReturnVal = &RetVal{[]byte(""), []byte(""), nil}
	}
}

func (r *RunCall) Returns(execBin string, stdout []byte, stderr []byte, err error) {
	r.initCall(execBin)
	call := r.Calls[execBin]
	call.ReturnVal = &RetVal{stdout, stderr, err}
}

func (r *RunCall) CallCount(execBin string) int {
	r.initCall(execBin)
	call := r.Calls[execBin]
	return call.CallCount
}

func (r *RunCall) ReceivedArgs(execBin string) []string {
	r.initCall(execBin)
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

func (c *Cmd) Run(executable string, args ...string) ([]byte, []byte, error) {
	execPath := strings.Split(executable, "\\")
	execBin := execPath[len(execPath)-1]
	c.RunCalls.initCall(execBin)
	call := c.RunCalls.Calls[execBin]
	call.RecievedArgs = args
	call.CallCount++
	ret := call.ReturnVal
	return ret.Stdout, ret.Stderr, ret.Err
}

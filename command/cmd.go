package command

import (
	"bytes"
	"os/exec"
)

type Cmd struct {
}

func NewCmd() *Cmd {
	return &Cmd{}
}

func (c *Cmd) Run(executable string, args ...string) (string, string, error) {
	var command *exec.Cmd
	var outbuff, errbuff bytes.Buffer
	var err error

	command = exec.Command(executable, args...)
	command.Stdout = &outbuff
	command.Stderr = &errbuff
	err = command.Run()

	return outbuff.String(), errbuff.String(), err
}

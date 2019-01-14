package command

import (
	"bytes"
	"os/exec"
)

type Cmd struct {
}

// TODO: Inject stdout and stderr instead of returning a buffer
func NewCmd() *Cmd {
	return &Cmd{}
}

func (c *Cmd) Run(executable string, args ...string) ([]byte, []byte, error) {
	var command *exec.Cmd
	var outbuff, errbuff bytes.Buffer
	var err error

	command = exec.Command(executable, args...)
	command.Stdout = &outbuff
	command.Stderr = &errbuff
	err = command.Run()
	return outbuff.Bytes(), errbuff.Bytes(), err
}

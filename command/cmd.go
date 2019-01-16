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

// TODO: Inject stdout and stderr instead of returning a buffer
func (c *Cmd) Run(executable string, args ...string) ([]byte, []byte, error) {
	var command *exec.Cmd
	var outbuff, errbuff bytes.Buffer
	var err error

	command = exec.Command(executable, args...)
	command.Stdout = &outbuff
	command.Stderr = &errbuff
	err = command.Run()

	// TODO: wrap error returned with executable and args
	return outbuff.Bytes(), errbuff.Bytes(), err
}

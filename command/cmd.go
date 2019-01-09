package command

import (
	"os/exec"
)

type Cmd struct {
}

func NewCmd() Cmd {
	return Cmd{}
}

func (c Cmd) Run(executable string, args ...string) error {
	return exec.Command(executable, args...).Run()
}

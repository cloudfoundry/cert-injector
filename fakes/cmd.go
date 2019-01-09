package fakes

type Cmd struct {
	RunCall struct {
		CallCount int
		Receives  struct {
			Executable string
			Args       []string
		}
		Returns struct {
			Error error
		}
	}
}

func (c *Cmd) Run(executable string, args ...string) error {
	c.RunCall.CallCount++
	c.RunCall.Receives.Executable = executable
	c.RunCall.Receives.Args = args
	return c.RunCall.Returns.Error
}

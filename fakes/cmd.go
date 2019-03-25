package fakes

type Cmd struct {
	RunCall struct {
		CallCount int
		Receives  []RunCallReceive
		Returns   []RunCallReturn
		OnCall    []RunCallOnCall
	}
}

type RunCallOnCall func(executable string, args ...string) (string, string, error)

type RunCallReceive struct {
	Executable string
	Args       []string
}

type RunCallReturn struct {
	Stdout string
	Stderr string
	Error  error
}

func (c *Cmd) Run(executable string, args ...string) (string, string, error) {
	c.RunCall.CallCount++

	c.RunCall.Receives = append(c.RunCall.Receives, RunCallReceive{
		Executable: executable,
		Args:       args,
	})

	if onCall := c.RunCall.OnCall[c.RunCall.CallCount-1]; onCall != nil {
		return onCall(executable, args...)
	}

	if len(c.RunCall.Returns) < c.RunCall.CallCount {
		return "", "", nil
	}

	return c.RunCall.Returns[c.RunCall.CallCount-1].Stdout, c.RunCall.Returns[c.RunCall.CallCount-1].Stderr, c.RunCall.Returns[c.RunCall.CallCount-1].Error
}

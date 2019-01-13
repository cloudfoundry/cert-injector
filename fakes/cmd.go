package fakes

type Cmd struct {
	RunCall struct {
		CallCount int
		Receives  []RunCallReceive
		Returns   []RunCallReturn
	}
}

type RunCallReceive struct {
	Executable string
	Args       []string
}

type RunCallReturn struct {
	Stdout []byte
	Stderr []byte
	Error  error
}

func (c *Cmd) Run(executable string, args ...string) ([]byte, []byte, error) {
	c.RunCall.CallCount++

	c.RunCall.Receives = append(c.RunCall.Receives, RunCallReceive{
		Executable: executable,
		Args:       args,
	})

	if len(c.RunCall.Returns) < c.RunCall.CallCount {
		return nil, nil, nil
	}

	return c.RunCall.Returns[c.RunCall.CallCount-1].Stdout, c.RunCall.Returns[c.RunCall.CallCount-1].Stderr, c.RunCall.Returns[c.RunCall.CallCount-1].Error
}

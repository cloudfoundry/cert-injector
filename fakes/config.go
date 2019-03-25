package fakes

type Config struct {
	WriteCall struct {
		CallCount int
		Receives  []WriteCallReceive
		Returns   []WriteCallReturn
	}
}

type WriteCallReceive struct {
	BundleDir   string
	GrootOutput string
	CertData    string
}

type WriteCallReturn struct {
	Error error
}

func (c *Config) Write(bundleDir string, grootOutput string, certData string) error {
	c.WriteCall.CallCount++

	c.WriteCall.Receives = append(c.WriteCall.Receives, WriteCallReceive{
		BundleDir:   bundleDir,
		GrootOutput: grootOutput,
		CertData:    certData,
	})

	if len(c.WriteCall.Returns) < c.WriteCall.CallCount {
		return nil
	}

	return c.WriteCall.Returns[c.WriteCall.CallCount-1].Error
}

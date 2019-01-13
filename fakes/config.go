package fakes

type Config struct {
	WriteCall struct {
		CallCount int
		Receives  struct {
			BundleDir   string
			GrootOutput []byte
			CertData    []byte
		}
		Returns struct {
			Error error
		}
	}
}

func (c *Config) Write(bundleDir string, grootOutput []byte, certData []byte) error {
	c.WriteCall.CallCount++
	c.WriteCall.Receives.BundleDir = bundleDir
	c.WriteCall.Receives.GrootOutput = grootOutput
	c.WriteCall.Receives.CertData = certData

	return c.WriteCall.Returns.Error
}

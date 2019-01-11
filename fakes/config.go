package fakes

type Config struct {
	WriteCall struct {
		CallCount int
		Receives  struct {
			BundleDir string
			CertData  []byte
		}
		Returns struct {
			Error error
		}
	}
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) Write(bundleDir string, certData []byte) error {
	c.WriteCall.CallCount++
	c.WriteCall.Receives.BundleDir = bundleDir
	c.WriteCall.Receives.CertData = certData

	return c.WriteCall.Returns.Error
}

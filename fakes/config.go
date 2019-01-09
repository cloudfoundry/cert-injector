package fakes

type Config struct {
	WriteCall struct {
		CallCount int
		Receives  struct {
			CertData []byte
		}
		Returns struct {
			Error error
		}
	}
}

func (c *Config) Write(certData []byte) error {
	c.WriteCall.CallCount++
	c.WriteCall.Receives.CertData = certData

	return c.WriteCall.Returns.Error
}

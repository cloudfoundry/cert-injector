package container_test

import (
	"certificate-injector/container"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var (
		conf     container.Config
		certData []byte
	)
	BeforeEach(func() {
		conf = container.NewConfig()
		certData = []byte("really-a-cert")
	})

	It("encodes a script to import the certificates and writes it to config.json", func() {
		err := conf.Write(certData)
		Expect(err).NotTo(HaveOccurred())

		data, err := ioutil.ReadFile("config.json")
		Expect(err).NotTo(HaveOccurred())

		cont := container.ConfigJSON{}
		json.Unmarshal(data, &cont)
		Expect(cont.Process.Cwd).To(Equal("C:\\"))

		decoded, err := base64.StdEncoding.DecodeString(cont.Process.Args[2])
		Expect(err).NotTo(HaveOccurred())
		Expect(string(decoded)).To(Equal(fmt.Sprintf(container.ImportCertificatePs, string(certData))))
	})
})

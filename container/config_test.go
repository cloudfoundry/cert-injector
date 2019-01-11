package container_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cert-injector/container"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var (
		bundleDir string
		certData  []byte

		conf container.Config
	)
	BeforeEach(func() {
		bundleDir = os.TempDir()
		certData = []byte("really-a-cert")

		conf = container.NewConfig()
	})

	It("encodes a script to import the certificates and writes it to config.json", func() {
		err := conf.Write(bundleDir, certData)
		Expect(err).NotTo(HaveOccurred())

		data, err := ioutil.ReadFile(filepath.Join(bundleDir, "config.json"))
		Expect(err).NotTo(HaveOccurred())

		cont := container.ConfigJSON{}
		json.Unmarshal(data, &cont)
		Expect(cont.Process.Cwd).To(Equal("C:\\"))

		decoded, err := base64.StdEncoding.DecodeString(cont.Process.Args[2])
		Expect(err).NotTo(HaveOccurred())
		Expect(string(decoded)).To(Equal(fmt.Sprintf(container.ImportCertificatePs, string(certData))))
	})
})

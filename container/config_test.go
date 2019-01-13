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
	oci "github.com/opencontainers/runtime-spec/specs-go"
)

var _ = Describe("Config", func() {
	var (
		bundleDir string
		certData  []byte
		path      string

		conf container.Config
	)
	BeforeEach(func() {
		bundleDir = os.TempDir()
		certData = []byte("really-a-cert")
		path = filepath.Join(bundleDir, "config.json")

		conf = container.NewConfig()
	})

	AfterEach(func() {
		Expect(os.Remove(path)).NotTo(HaveOccurred())
	})

	It("encodes a script to import the certificates and writes it to config.json", func() {
		err := conf.Write(bundleDir, certData)
		Expect(err).NotTo(HaveOccurred())

		data, err := ioutil.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())

		cont := oci.Spec{}
		json.Unmarshal(data, &cont)
		Expect(cont.Process.Cwd).To(Equal("C:\\"))

		decoded, err := base64.StdEncoding.DecodeString(cont.Process.Args[2])
		Expect(err).NotTo(HaveOccurred())
		Expect(string(decoded)).To(Equal(fmt.Sprintf(container.ImportCertificatePs, string(certData))))
	})
})

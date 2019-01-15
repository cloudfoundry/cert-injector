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
		bundleDir   string
		certData    string
		grootOutput []byte
		path        string

		conf container.Config
	)
	BeforeEach(func() {
		bundleDir = os.TempDir()
		certData = "cert-base64-encoded"
		grootOutput = []byte(`{"ociVersion": "2.2.2"}`)
		path = filepath.Join(bundleDir, "config.json")

		conf = container.NewConfig()
	})

	AfterEach(func() {
		Expect(os.RemoveAll(path)).NotTo(HaveOccurred())
	})

	It("encodes a script to import the certificates and writes it to config.json", func() {
		err := conf.Write(bundleDir, grootOutput, certData)
		Expect(err).NotTo(HaveOccurred())

		data, err := ioutil.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())

		cont := oci.Spec{}
		json.Unmarshal(data, &cont)
		Expect(cont.Version).To(Equal("2.2.2"))
		Expect(cont.Process.Cwd).To(Equal("C:\\"))

		decoded, err := base64.StdEncoding.DecodeString(cont.Process.Args[2])
		Expect(err).NotTo(HaveOccurred())
		Expect(string(decoded)).To(Equal(fmt.Sprintf(container.ImportCertificatePs, certData)))
	})

	Context("when the groot output is invalid json", func() {
		It("returns  helpful error message", func() {
			err := conf.Write(bundleDir, []byte("$$$"), certData)
			Expect(err).To(MatchError("json unmarshal groot output: invalid character '$' looking for beginning of value"))
		})
	})
})

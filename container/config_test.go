package container_test

import (
	"encoding/json"
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
		bundleDir     string
		certDirectory string
		grootOutput   []byte
		path          string

		conf container.Config
	)
	BeforeEach(func() {
		bundleDir = os.TempDir()
		certDirectory = "some-directory-containing-certs"
		grootOutput = []byte(`{"ociVersion": "2.2.2"}`)
		path = filepath.Join(bundleDir, "config.json")

		conf = container.NewConfig()
	})

	AfterEach(func() {
		Expect(os.RemoveAll(path)).NotTo(HaveOccurred())
	})

	It("the config.json contains a process spec to import the certificates, and bind-mounts the certificates into the container", func() {
		err := conf.Write(bundleDir, grootOutput, certDirectory)
		Expect(err).NotTo(HaveOccurred())

		data, err := ioutil.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())

		cont := oci.Spec{}
		json.Unmarshal(data, &cont)
		Expect(cont.Version).To(Equal("2.2.2"))
		Expect(cont.Process.Cwd).To(Equal("C:\\"))
		Expect(cont.Process.Args).To(ConsistOf("powershell.exe", "-Command", container.ImportCertificatePs))
	})

	Context("when the groot output is invalid json", func() {
		It("returns  helpful error message", func() {
			err := conf.Write(bundleDir, []byte("$$$"), certDirectory)
			Expect(err).To(MatchError("json unmarshal groot output: invalid character '$' looking for beginning of value"))
		})
	})
})

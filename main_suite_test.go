package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var certInjectorBin string

func TestCertificateInjector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CertificateInjector Suite")
}

/*
var _ = BeforeSuite(func() {
	certInjectorBin, err = gexec.Build("certificate-injector")
	Expect(err).ToNot(HaveOccurred())
})
*/

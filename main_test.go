package main_test

import (
	"errors"
	"fmt"

	. "code.cloudfoundry.org/cert-injector"
	"code.cloudfoundry.org/cert-injector/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("cert-injector", func() {
	var (
		fakeCmd    *fakes.Cmd
		fakeConfig *fakes.Config

		args []string
	)

	BeforeEach(func() {
		fakeCmd = fakes.NewCmd()
		fakeConfig = fakes.NewConfig()

		args = []string{"cert-injector.exe", "", "fakes/really-has-certs.crt", "first-image-uri"}
	})

	It("calls hydrator to remove the custom layer", func() {
		err := Run(args, fakeCmd, fakeConfig)
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeCmd.RunCalls.CallCount("hydrate.exe")).NotTo(Equal(0))
		Expect(fakeCmd.RunCalls.ReceivedArgs("hydrate.exe")).To(ContainElement("remove-layer"))
		Expect(fakeCmd.RunCalls.ReceivedArgs("hydrate.exe")).To(ContainElement("first-image-uri"))
	})

	Context("when hydrator fails to remove the custom layer", func() {
		BeforeEach(func() {
			fakeCmd.RunCalls.Returns("hydrate.exe", errors.New("hydrator is unhappy"))
		})

		It("should return a helpful error", func() {
			err := Run(args, fakeCmd, fakeConfig)
			Expect(err).To(MatchError("hydrate.exe remove-layer failed: hydrator is unhappy\n"))
		})
	})

	Describe("cert_file", func() {
		Context("when the cert_file does not exist", func() {
			BeforeEach(func() {
				args = []string{"cert-injector.exe", "", "not-a-real-file.crt", "first-image-uri"}
			})

			It("returns a helpful error", func() {
				err := Run(args, fakeCmd, fakeConfig)
				Expect(err).To(MatchError("Failed to read cert_file: open not-a-real-file.crt: no such file or directory"))
			})
		})

		Context("when there are no trusted certs to inject", func() {
			BeforeEach(func() {
				args = []string{"cert-injector.exe", "", "fakes/empty.crt", "first-image-uri"}
			})

			It("does not check other arguments and exits successfully", func() {
				err := Run(args, fakeCmd, fakeConfig)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("config.json", func() {
		It("creates a config for the container", func() {
			err := Run(args, fakeCmd, fakeConfig)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeConfig.WriteCall.CallCount).To(Equal(1))
			Expect(string(fakeConfig.WriteCall.Receives.CertData)).To(ContainSubstring("this-is-a-cert"))
		})

		Context("when it fails to create a config", func() {
			BeforeEach(func() {
				fakeConfig.WriteCall.Returns.Error = errors.New("banana")
			})

			It("returns a helpful error message", func() {
				err := Run(args, fakeCmd, fakeConfig)
				Expect(err).To(MatchError("Write container config failed: banana"))
			})
		})
	})

	Describe("diff-exporter is called", func() {
		It("with -bundlePath containing a valid config.json", func() {
			/* For the config.json to be valid, it should contain the script to import cert */
			err := Run(args, fakeCmd, fakeConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCmd.RunCalls.CallCount("diff-exporter.exe")).To(Equal(1))
			Fail("TODO: INCOMPLETE")
		})
	})

	It("hydrator is called with add-layer command exactly once", func() {
	})

	Context("When the hydrator is called with add-layer command", func() {
		It("It was called with a layer that contains the required certificate", func() {
		})
	})

	Context("when called with incorrect arguments", func() {
		It("returns a helpful error message with usage", func() {
			err := Run([]string{"cert-injector.exe"}, fakeCmd, fakeConfig)
			Expect(err).To(MatchError(fmt.Sprintf("usage: %s <driver_store> <cert_file> <image_uri>...\n", args[0])))
		})
	})
})

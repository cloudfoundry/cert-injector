package main_test

import (
	. "certificate-injector"
	fakes "certificate-injector/fakes"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("certificate-injector", func() {
	var (
		fakeImage  *fakes.Image
		fakeCmd    *fakes.Cmd
		fakeConfig *fakes.Config

		args []string
	)

	BeforeEach(func() {
		fakeImage = &fakes.Image{}
		fakeCmd = &fakes.Cmd{}
		fakeConfig = &fakes.Config{}

		args = []string{"certificate-injector.exe", "", "fakes/really-has-certs.crt", "first-image-uri"}
	})

	It("calls hydrator to remove the custom layer", func() {
		err := Run(args, fakeImage, fakeCmd, fakeConfig)
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeCmd.RunCall.CallCount).To(Equal(1))
		Expect(fakeCmd.RunCall.Receives.Executable).To(ContainSubstring("hydrate.exe"))
		Expect(fakeCmd.RunCall.Receives.Args).To(ContainElement("first-image-uri"))
	})

	Context("when hydrator fails to remove the custom layer", func() {
		BeforeEach(func() {
			fakeCmd.RunCall.Returns.Error = errors.New("hydrator is unhappy")
		})

		It("should return a helpful error", func() {
			err := Run(args, fakeImage, fakeCmd, fakeConfig)
			Expect(err).To(MatchError("hydrate.exe remove-layer failed: hydrator is unhappy\n"))
		})
	})

	Describe("cert_file", func() {
		Context("when the cert_file does not exist", func() {
			BeforeEach(func() {
				args = []string{"certificate-injector.exe", "", "not-a-real-file.crt", "first-image-uri"}
			})

			It("returns a helpful error", func() {
				err := Run(args, fakeImage, fakeCmd, fakeConfig)
				Expect(err).To(MatchError("Failed to read cert_file: open not-a-real-file.crt: no such file or directory"))
			})
		})

		Context("when there are no trusted certs to inject", func() {
			BeforeEach(func() {
				args = []string{"certificate-injector.exe", "", "fakes/empty.crt", "first-image-uri"}
			})

			It("does not check other arguments and exits successfully", func() {
				err := Run(args, fakeImage, fakeCmd, fakeConfig)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeImage.ContainsHydratorAnnotationCall.CallCount).To(Equal(0))
			})
		})
	})

	Describe("config.json", func() {
		It("creates a config for the container", func() {
			err := Run(args, fakeImage, fakeCmd, fakeConfig)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeConfig.WriteCall.CallCount).To(Equal(1))
			Expect(string(fakeConfig.WriteCall.Receives.CertData)).To(ContainSubstring("this-is-a-cert"))
		})

		Context("when it fails to create a config", func() {
			BeforeEach(func() {
				fakeConfig.WriteCall.Returns.Error = errors.New("banana")
			})

			It("returns a helpful error message", func() {
				err := Run(args, fakeImage, fakeCmd, fakeConfig)
				Expect(err).To(MatchError("Write container config failed: banana"))
			})
		})
	})

	Context("when called with incorrect arguments", func() {
		It("returns a helpful error message with usage", func() {
			err := Run([]string{"certificate-injector.exe"}, fakeImage, fakeCmd, fakeConfig)
			Expect(err).To(MatchError(fmt.Sprintf("usage: %s <driver_store> <cert_file> <image_uri>...\n", args[0])))
		})
	})
})

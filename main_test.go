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

		driverStore string
		ociImageUri string
		grootOutput []byte
		args        []string
	)

	BeforeEach(func() {
		fakeCmd = &fakes.Cmd{}
		fakeConfig = &fakes.Config{}

		driverStore = "some-driver-store"
		ociImageUri = "first-image-uri"
		grootOutput = []byte("gibberish")
		args = []string{"cert-injector.exe", driverStore, "fakes/really-has-certs.crt", ociImageUri}

		fakeCmd.RunCall.Returns = make([]fakes.RunCallReturn, 5)
		fakeCmd.RunCall.Returns[1].Stdout = grootOutput
	})

	It("replaces custom layers with a new layer with new certificates", func() {
		err := Run(args, fakeCmd, fakeConfig)
		Expect(err).NotTo(HaveOccurred())

		By("calling hydrator to remove the current custom layers")
		Expect(fakeCmd.RunCall.Receives[0].Executable).To(ContainSubstring("hydrate.exe"))
		Expect(fakeCmd.RunCall.Receives[0].Args).To(ConsistOf("remove-layer", "-ociImage", ociImageUri))

		By("calling groot to create a volume")
		Expect(fakeCmd.RunCall.Receives[1].Executable).To(ContainSubstring("groot.exe"))
		Expect(fakeCmd.RunCall.Receives[1].Args).To(ConsistOf("--driver-store", driverStore, "create", ociImageUri))

		By("creating a bundle directory and container config")
		Expect(fakeConfig.WriteCall.CallCount).To(Equal(1))
		Expect(fakeConfig.WriteCall.Receives.BundleDir).To(ContainSubstring("layer"))
		Expect(fakeConfig.WriteCall.Receives.GrootOutput).To(Equal(grootOutput))
		Expect(string(fakeConfig.WriteCall.Receives.CertData)).To(ContainSubstring("this-is-a-cert"))

		By("calling winc to create a container")
		Expect(fakeCmd.RunCall.Receives[2].Executable).To(ContainSubstring("winc.exe"))
		Expect(fakeCmd.RunCall.Receives[2].Args).To(ConsistOf("run", "-b", ContainSubstring("layer"), ContainSubstring("layer")))

		By("calling diff-exporter to export the top layer")
		Expect(fakeCmd.RunCall.Receives[3].Executable).To(ContainSubstring("diff-exporter.exe"))
		Expect(fakeCmd.RunCall.Receives[3].Args).To(ConsistOf("-outputFile", ContainSubstring("diff-output"), "-containerId", ContainSubstring("layer"), "-bundlePath", ContainSubstring("layer")))
	})

	Describe("error cases", func() {
		Context("when cert-injector is called with incorrect arguments", func() {
			It("prints the usage", func() {
				err := Run([]string{"cert-injector.exe"}, fakeCmd, fakeConfig)
				Expect(err).To(MatchError(fmt.Sprintf("usage: %s <driver_store> <cert_file> <image_uri>...\n", args[0])))
			})
		})

		Context("when the cert_file does not exist", func() {
			BeforeEach(func() {
				args = []string{"cert-injector.exe", "", "not-a-real-file.crt", ""}
			})

			It("returns a helpful error", func() {
				err := Run(args, fakeCmd, fakeConfig)
				Expect(err).To(MatchError("Failed to read cert_file: open not-a-real-file.crt: no such file or directory"))
			})
		})

		Context("when there are no trusted certs to inject", func() {
			BeforeEach(func() {
				args = []string{"cert-injector.exe", "", "fakes/empty.crt", ""}
			})

			It("does not check other arguments and exits successfully", func() {
				err := Run(args, fakeCmd, fakeConfig)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when hydrator fails to remove the custom layer", func() {
			BeforeEach(func() {
				fakeCmd.RunCall.Returns[0].Error = errors.New("hydrator is unhappy")
			})

			It("should return a helpful error", func() {
				err := Run(args, fakeCmd, fakeConfig)
				Expect(err).To(MatchError("hydrate.exe remove-layer failed: hydrator is unhappy\n"))
			})
		})

		Context("when groot fails to create a volume", func() {
			BeforeEach(func() {
				fakeCmd.RunCall.Returns[1].Error = errors.New("groot is unhappy")
			})
			It("returns a helpful error message", func() {
				err := Run(args, fakeCmd, fakeConfig)
				Expect(err).To(MatchError("groot create failed: groot is unhappy"))
			})
		})

		Context("when config write fails", func() {
			BeforeEach(func() {
				fakeConfig.WriteCall.Returns.Error = errors.New("banana")
			})

			It("returns a helpful error message", func() {
				err := Run(args, fakeCmd, fakeConfig)
				Expect(err).To(MatchError("Write container config failed: banana"))
			})
		})

		Context("when winc fails to create a container", func() {
			BeforeEach(func() {
				fakeCmd.RunCall.Returns[2].Error = errors.New("winc is unhappy")
			})

			It("returns a helpful error message", func() {
				err := Run(args, fakeCmd, fakeConfig)
				Expect(err).To(MatchError("winc run failed: winc is unhappy"))
			})
		})

		Context("when diff-exporter fails to export the top layer", func() {
			BeforeEach(func() {
				fakeCmd.RunCall.Returns[3].Error = errors.New("diff-exporter is unhappy")
			})
			It("returns a helpful error message", func() {
				err := Run(args, fakeCmd, fakeConfig)
				Expect(err).To(MatchError("diff-exporter failed: diff-exporter is unhappy"))
			})
		})
	})
})

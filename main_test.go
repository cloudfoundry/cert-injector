package main_test

import (
	"errors"
	"fmt"
	"os"

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
		args        []string
	)

	BeforeEach(func() {
		fakeCmd = fakes.NewCmd()
		fakeConfig = fakes.NewConfig()
		driverStore = "some-driver-store"
		ociImageUri = "first-image-uri"
		args = []string{"cert-injector.exe", driverStore, "fakes/really-has-certs.crt", ociImageUri}
	})

	Context("when cert-injector is called with incorrect arguments", func() {
		It("prints the usage", func() {
			err := Run([]string{"cert-injector.exe"}, fakeCmd, fakeConfig)
			Expect(err).To(MatchError(fmt.Sprintf("usage: %s <driver_store> <cert_file> <image_uri>...\n", args[0])))
		})
	})

	Describe("hydrator", func() {
		It("calls hydrator to remove any custom layers", func() {
			err := Run(args, fakeCmd, fakeConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCmd.RunCalls.CallCount("hydrate.exe")).NotTo(Equal(0))
			Expect(fakeCmd.RunCalls.ReceivedArgs("hydrate.exe")).To(ConsistOf("remove-layer", "-ociImage", ociImageUri))
		})

		Context("when hydrator fails to remove the custom layer", func() {
			BeforeEach(func() {
				fakeCmd.RunCalls.Returns("hydrate.exe", nil, nil, errors.New("hydrator is unhappy"))
			})

			It("should return a helpful error", func() {
				err := Run(args, fakeCmd, fakeConfig)
				Expect(err).To(MatchError("hydrate.exe remove-layer failed: hydrator is unhappy\n"))
			})
		})
	})

	Describe("cert_file", func() {
		Context("when the file does not exist", func() {
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
	})

	Describe("bundle directory and config", func() {
		It("creates a temp directory to write the container cnonfig", func() {
			err := Run(args, fakeCmd, fakeConfig)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeConfig.WriteCall.CallCount).To(Equal(1))
			bundleDir := fakeConfig.WriteCall.Receives.BundleDir
			Expect(bundleDir).To(ContainSubstring("layer"))
			Expect(string(fakeConfig.WriteCall.Receives.CertData)).To(ContainSubstring("this-is-a-cert"))

			By("deleting the bundle directory at the end")
			_, err = os.Stat(bundleDir)
			Expect(err).NotTo(Succeed())
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

	Describe("groot", func() {
		It("creates a volume with the appropriate arguments", func() {
			err := Run(args, fakeCmd, fakeConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCmd.RunCalls.CallCount("groot.exe")).To(Equal(1))

			receivedArgs := fakeCmd.RunCalls.ReceivedArgs("groot.exe")
			Expect(receivedArgs).To(ConsistOf("--driver-store", driverStore, "create", ociImageUri))
		})

		Context("when winc fails", func() {
			BeforeEach(func() {
				fakeCmd.RunCalls.Returns("groot.exe", nil, nil, errors.New("banana"))
			})
			It("returns a helpful error message", func() {
				err := Run(args, fakeCmd, fakeConfig)
				Expect(err).To(MatchError("groot create failed: banana"))
			})
		})
	})

	Describe("winc", func() {
		It("creates a container with the appropriate arguments", func() {
			err := Run(args, fakeCmd, fakeConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCmd.RunCalls.CallCount("winc.exe")).To(Equal(1))

			receivedArgs := fakeCmd.RunCalls.ReceivedArgs("winc.exe")
			Expect(receivedArgs).To(ConsistOf("run", "-b", ContainSubstring("layer"), ContainSubstring("layer")))
		})

		Context("when winc fails", func() {
			BeforeEach(func() {
				fakeCmd.RunCalls.Returns("winc.exe", nil, nil, errors.New("banana"))
			})
			It("returns a helpful error message", func() {
				err := Run(args, fakeCmd, fakeConfig)
				Expect(err).To(MatchError("winc run failed: banana"))
			})
		})
	})

	Describe("diff-exporter", func() {
		It("exports the top layer with the appropriate arguments", func() {
			err := Run(args, fakeCmd, fakeConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCmd.RunCalls.CallCount("diff-exporter.exe")).To(Equal(1))

			receivedArgs := fakeCmd.RunCalls.ReceivedArgs("diff-exporter.exe")
			Expect(receivedArgs).To(ConsistOf("-outputFile", ContainSubstring("diff-output"), "-containerId", ContainSubstring("layer"), "-bundlePath", ContainSubstring("layer")))
		})

		Context("when diff-exporter fails", func() {
			BeforeEach(func() {
				fakeCmd.RunCalls.Returns("diff-exporter.exe", nil, nil, errors.New("banana"))
			})
			It("returns a helpful error message", func() {
				err := Run(args, fakeCmd, fakeConfig)
				Expect(err).To(MatchError("diff-exporter failed: banana"))
			})
		})
	})

	Describe("hydrator", func() {
		PIt("hydrator is called with add-layer command exactly once", func() {
		})

		Context("When the hydrator is called with add-layer command", func() {
			PIt("It was called with a layer that contains the required certificate", func() {
			})
		})
	})

})

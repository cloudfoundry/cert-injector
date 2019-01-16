package main_test

import (
	"errors"
	"fmt"
	"io/ioutil"

	. "code.cloudfoundry.org/cert-injector"
	"code.cloudfoundry.org/cert-injector/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("cert-injector", func() {
	var (
		fakeCmd    *fakes.Cmd
		fakeConfig *fakes.Config
		stdout     *fakes.Logger
		stderr     *fakes.Logger

		driverStore string
		certData    string
		ociImageUri string
		grootOutput []byte
		args        []string
		layerTgz    string
	)

	BeforeEach(func() {
		fakeCmd = &fakes.Cmd{}
		fakeConfig = &fakes.Config{}
		stdout = &fakes.Logger{}
		stderr = &fakes.Logger{}

		driverStore = "some-driver-store"
		certData = "cert-data-base64-encoded"
		ociImageUri = "oci:///first-image-uri"
		grootOutput = []byte("gibberish")
		args = []string{"cert-injector.exe", driverStore, certData, ociImageUri}

		fakeCmd.RunCall.OnCall = make([]fakes.RunCallOnCall, 20)
		fakeCmd.RunCall.Returns = make([]fakes.RunCallReturn, 20)
		fakeCmd.RunCall.Returns[1].Stdout = grootOutput
		fakeCmd.RunCall.OnCall[3] = func(executable string, args ...string) ([]byte, []byte, error) {
			Expect(executable).To(ContainSubstring("diff-exporter.exe"))
			layerTgz = args[1]
			Expect(layerTgz).To(ContainSubstring("diff-output"))
			Expect(ioutil.WriteFile(layerTgz, []byte("some-tar-data"), 0644)).To(Succeed())
			return nil, nil, nil
		}

		fakeConfig.WriteCall.Returns = make([]fakes.WriteCallReturn, 2)
	})

	It("replaces custom layers with a new layer with new certificates", func() {
		err := Run(args, fakeCmd, fakeConfig, stdout, stderr)
		Expect(err).NotTo(HaveOccurred())

		By("calling hydrator to remove the old layer")
		Expect(fakeCmd.RunCall.Receives[0].Executable).To(ContainSubstring("hydrate.exe"))
		Expect(fakeCmd.RunCall.Receives[0].Args).To(ConsistOf("remove-layer", "-ociImage", ociImageUri))

		By("calling groot to create a volume")
		Expect(fakeCmd.RunCall.Receives[1].Executable).To(ContainSubstring("groot.exe"))
		Expect(fakeCmd.RunCall.Receives[1].Args).To(ConsistOf("--driver-store", driverStore, "create", ociImageUri, ContainSubstring("layer")))

		By("creating a bundle directory and container config")
		Expect(fakeConfig.WriteCall.CallCount).To(Equal(1))
		Expect(fakeConfig.WriteCall.Receives[0].BundleDir).To(ContainSubstring("layer"))
		Expect(fakeConfig.WriteCall.Receives[0].GrootOutput).To(Equal(grootOutput))
		Expect(fakeConfig.WriteCall.Receives[0].CertData).To(Equal(certData))

		By("calling winc to create a container")
		Expect(fakeCmd.RunCall.Receives[2].Executable).To(ContainSubstring("winc.exe"))
		Expect(fakeCmd.RunCall.Receives[2].Args).To(ConsistOf("run", "-b", ContainSubstring("layer"), ContainSubstring("layer")))

		By("calling diff-exporter to export the top layer")
		Expect(fakeCmd.RunCall.Receives[3].Executable).To(ContainSubstring("diff-exporter.exe"))
		Expect(fakeCmd.RunCall.Receives[3].Args).To(ConsistOf("-outputFile", ContainSubstring("diff-output"), "-containerId", ContainSubstring("layer"), "-bundlePath", ContainSubstring("layer")))

		By("calling hydrator to add the new layer")
		Expect(fakeCmd.RunCall.Receives[4].Executable).To(ContainSubstring("hydrate.exe"))
		Expect(fakeCmd.RunCall.Receives[4].Args).To(ConsistOf("add-layer", "-ociImage", ociImageUri, "-layer", ContainSubstring("diff-output")))

		By("calling groot to delete the volume")
		Expect(fakeCmd.RunCall.Receives[5].Executable).To(ContainSubstring("groot.exe"))
		Expect(fakeCmd.RunCall.Receives[5].Args).To(ConsistOf("--driver-store", driverStore, "delete", ContainSubstring("layer")))

		By("checking bundle dir is gone")
		Expect(fakeConfig.WriteCall.Receives[0].BundleDir).NotTo(BeAnExistingFile())

		By("checking layer.tgz is gone")
		Expect(layerTgz).NotTo(BeAnExistingFile())
	})

	Context("when there are multiple oci image uris", func() {
		var (
			ociImageUri2 string
			layerTgz2    string
		)

		BeforeEach(func() {
			ociImageUri2 = "the-other-image-uri"
			args = append(args, ociImageUri2)

			fakeCmd.RunCall.Returns[7].Stdout = grootOutput
			fakeCmd.RunCall.OnCall[3] = nil

			fakeCmd.RunCall.OnCall[9] = func(executable string, args ...string) ([]byte, []byte, error) {
				Expect(executable).To(ContainSubstring("diff-exporter.exe"))
				layerTgz2 = args[1]
				Expect(layerTgz2).To(ContainSubstring("diff-output"))
				Expect(ioutil.WriteFile(layerTgz2, []byte("some-tar-data"), 0644)).To(Succeed())
				return nil, nil, nil
			}
		})

		It("does the loop for each one", func() {
			err := Run(args, fakeCmd, fakeConfig, stdout, stderr)
			Expect(err).NotTo(HaveOccurred())

			By("calling hydrator to remove the old layer twice")
			Expect(fakeCmd.RunCall.Receives[1].Executable).To(ContainSubstring("hydrate.exe"))
			Expect(fakeCmd.RunCall.Receives[1].Args).To(ConsistOf("remove-layer", "-ociImage", ociImageUri2))

			By("calling groot to create a volume twice")
			Expect(fakeCmd.RunCall.Receives[7].Executable).To(ContainSubstring("groot.exe"))
			Expect(fakeCmd.RunCall.Receives[7].Args).To(ConsistOf("--driver-store", driverStore, "create", ociImageUri2, ContainSubstring("layer")))

			By("creating a bundle directory and container config twice")
			Expect(fakeConfig.WriteCall.Receives[1].BundleDir).To(ContainSubstring("layer"))
			Expect(fakeConfig.WriteCall.Receives[1].GrootOutput).To(Equal(grootOutput))
			Expect(fakeConfig.WriteCall.Receives[1].CertData).To(Equal(certData))

			By("calling winc to create a container twice")
			Expect(fakeCmd.RunCall.Receives[8].Executable).To(ContainSubstring("winc.exe"))
			Expect(fakeCmd.RunCall.Receives[8].Args).To(ConsistOf("run", "-b", ContainSubstring("layer"), ContainSubstring("layer")))

			By("calling diff-exporter to export the top layer twice")
			Expect(fakeCmd.RunCall.Receives[9].Executable).To(ContainSubstring("diff-exporter.exe"))
			Expect(fakeCmd.RunCall.Receives[9].Args).To(ConsistOf("-outputFile", ContainSubstring("diff-output"), "-containerId", ContainSubstring("layer"), "-bundlePath", ContainSubstring("layer")))

			By("calling hydrator to add the new layer twice")
			Expect(fakeCmd.RunCall.Receives[10].Executable).To(ContainSubstring("hydrate.exe"))
			Expect(fakeCmd.RunCall.Receives[10].Args).To(ConsistOf("add-layer", "-ociImage", ociImageUri2, "-layer", ContainSubstring("diff-output")))

			By("calling groot to delete the volume twice")
			Expect(fakeCmd.RunCall.Receives[11].Executable).To(ContainSubstring("groot.exe"))
			Expect(fakeCmd.RunCall.Receives[11].Args).To(ConsistOf("--driver-store", driverStore, "delete", ContainSubstring("layer")))

			By("checking bundle dir is gone")
			Expect(fakeConfig.WriteCall.Receives[1].BundleDir).NotTo(BeAnExistingFile())

			By("checking layer.tgz is gone twice")
			Expect(layerTgz).NotTo(BeAnExistingFile())
		})
	})

	Describe("error cases", func() {
		BeforeEach(func() {
			fakeCmd.RunCall.OnCall[3] = nil
		})

		Context("when cert-injector is called with incorrect arguments", func() {
			It("prints the usage", func() {
				err := Run([]string{"cert-injector.exe"}, fakeCmd, fakeConfig, stdout, stderr)
				Expect(err).To(MatchError(fmt.Sprintf("usage: %s <driver_store> <cert_data> <image_uri>...\n", args[0])))
			})
		})

		Context("when hydrator fails to remove the custom layer", func() {
			BeforeEach(func() {
				fakeCmd.RunCall.Returns[0].Error = errors.New("hydrator is unhappy")
			})

			It("should return a helpful error", func() {
				err := Run(args, fakeCmd, fakeConfig, stdout, stderr)
				Expect(err).To(MatchError("hydrate.exe remove-layer -ociImage oci:///first-image-uri failed: hydrator is unhappy\n"))
			})
		})

		Context("when groot fails to create a volume", func() {
			BeforeEach(func() {
				fakeCmd.RunCall.Returns[1].Error = errors.New("groot is unhappy")
			})
			It("returns a helpful error message", func() {
				err := Run(args, fakeCmd, fakeConfig, stdout, stderr)
				Expect(err).To(MatchError("groot create failed: groot is unhappy"))
			})
		})

		Context("when config write fails", func() {
			BeforeEach(func() {
				fakeConfig.WriteCall.Returns[0].Error = errors.New("banana")
			})

			It("returns a helpful error message, deletes the bundle dir, and deletes the volume created by groot", func() {
				err := Run(args, fakeCmd, fakeConfig, stdout, stderr)
				Expect(err).To(MatchError("container config write failed: banana"))

				Expect(fakeConfig.WriteCall.Receives[0].BundleDir).NotTo(BeAnExistingFile())
				Expect(fakeCmd.RunCall.Receives[2].Executable).To(ContainSubstring("groot.exe"))
				Expect(fakeCmd.RunCall.Receives[2].Args).To(ConsistOf("--driver-store", driverStore, "delete", ContainSubstring("layer")))
			})
		})

		Context("when winc fails to create a container", func() {
			BeforeEach(func() {
				fakeCmd.RunCall.Returns[2].Error = errors.New("winc is unhappy")
			})

			It("returns a helpful error message, deletes the bundle dir, and deletes the volume created by groot", func() {
				err := Run(args, fakeCmd, fakeConfig, stdout, stderr)
				Expect(err).To(MatchError("winc run failed: winc is unhappy"))

				Expect(fakeConfig.WriteCall.Receives[0].BundleDir).NotTo(BeAnExistingFile())
				Expect(fakeCmd.RunCall.Receives[3].Executable).To(ContainSubstring("groot.exe"))
				Expect(fakeCmd.RunCall.Receives[3].Args).To(ConsistOf("--driver-store", driverStore, "delete", ContainSubstring("layer")))
			})
		})

		Context("when diff-exporter fails to export the top layer", func() {
			BeforeEach(func() {
				fakeCmd.RunCall.Returns[3].Error = errors.New("diff-exporter is unhappy")
			})
			It("returns a helpful error message, deletes the bundle dir, and deletes the volume created by groot", func() {
				err := Run(args, fakeCmd, fakeConfig, stdout, stderr)
				Expect(err).To(MatchError("diff-exporter failed exporting the layer: diff-exporter is unhappy"))

				Expect(fakeConfig.WriteCall.Receives[0].BundleDir).NotTo(BeAnExistingFile())
				Expect(fakeCmd.RunCall.Receives[4].Executable).To(ContainSubstring("groot.exe"))
				Expect(fakeCmd.RunCall.Receives[4].Args).To(ConsistOf("--driver-store", driverStore, "delete", ContainSubstring("layer")))
			})
		})

		Context("when hydrator fails to add the new layer", func() {
			BeforeEach(func() {
				fakeCmd.RunCall.OnCall[3] = func(executable string, args ...string) ([]byte, []byte, error) {
					Expect(executable).To(ContainSubstring("diff-exporter.exe"))
					layerTgz = args[1]
					Expect(layerTgz).To(ContainSubstring("diff-output"))
					Expect(ioutil.WriteFile(layerTgz, []byte("some-tar-data"), 0644)).To(Succeed())
					return nil, nil, nil
				}
				fakeCmd.RunCall.Returns[4].Error = errors.New("hydrate add-layer is unhappy")
			})

			It("should return a helpful error, deletes the bundle dir, deletes the volume created by groot, and deletes the exported layer.tgz", func() {
				err := Run(args, fakeCmd, fakeConfig, stdout, stderr)
				Expect(err).To(MatchError("hydrate add-layer failed: hydrate add-layer is unhappy"))

				Expect(fakeConfig.WriteCall.Receives[0].BundleDir).NotTo(BeAnExistingFile())
				Expect(fakeCmd.RunCall.Receives[5].Executable).To(ContainSubstring("groot.exe"))
				Expect(fakeCmd.RunCall.Receives[5].Args).To(ConsistOf("--driver-store", driverStore, "delete", ContainSubstring("layer")))
				Expect(layerTgz).NotTo(BeAnExistingFile())
			})
		})

		Context("when groot fails to delete a volume", func() {
			BeforeEach(func() {
				fakeCmd.RunCall.Returns[5].Error = errors.New("groot is unhappy")
			})

			It("logs a helpful error message, but does not error", func() {
				err := Run(args, fakeCmd, fakeConfig, stdout, stderr)
				Expect(err).NotTo(HaveOccurred())

				Expect(stdout.PrintlnCall.Receives[0].Args[0]).To(Equal("groot delete failed"))
			})
		})
	})
})

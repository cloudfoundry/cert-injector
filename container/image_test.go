package container_test

import (
	"certificate-injector/container"
	"certificate-injector/container/fakes"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

var _ = Describe("Image", func() {
	var (
		handler *fakes.Handler
		image   container.Image
	)

	BeforeEach(func() {
		handler = &fakes.Handler{}
		image = container.NewImage(handler)
	})

	Context("when the image contains the layerAdded annotation", func() {
		BeforeEach(func() {
			handler.ReadMetadataCall.Returns.Manifest = oci.Manifest{
				Annotations: map[string]string{
					"hydrator.layerAdded": "true",
				},
			}
		})

		It("checks if a layer contains a hydrator annotation", func() {
			annotated, err := image.ContainsHydratorAnnotation("imageUri")
			Expect(err).NotTo(HaveOccurred())
			Expect(annotated).To(BeTrue())
			Expect(handler.ReadMetadataCall.CallCount).To(Equal(1))
		})
	})

	Context("when the image does not have the layerAdded annotation", func() {
		BeforeEach(func() {
			handler.ReadMetadataCall.Returns.Manifest = oci.Manifest{
				Annotations: map[string]string{},
			}
		})

		It("returns false", func() {
			annotated, err := image.ContainsHydratorAnnotation("imageUri")
			Expect(err).NotTo(HaveOccurred())
			Expect(annotated).To(BeFalse())
			Expect(handler.ReadMetadataCall.CallCount).To(Equal(1))
		})
	})

	Context("when the handler fails to read metadata", func() {
		BeforeEach(func() {
			handler.ReadMetadataCall.Returns.Error = errors.New("banana")
		})

		It("returns a helpful error message", func() {
			annotated, err := image.ContainsHydratorAnnotation("imageUri")
			Expect(err).To(MatchError("Hydrator handler failed to read metadata: banana"))
			Expect(annotated).To(BeFalse())

			Expect(handler.ReadMetadataCall.CallCount).To(Equal(1))
		})
	})
})

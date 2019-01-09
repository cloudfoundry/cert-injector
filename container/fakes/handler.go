package fakes

import (
	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

type Handler struct {
	ReadMetadataCall struct {
		CallCount int
		Returns   struct {
			Manifest oci.Manifest
			Image    oci.Image
			Error    error
		}
	}
}

func (h *Handler) ReadMetadata() (oci.Manifest, oci.Image, error) {
	h.ReadMetadataCall.CallCount++

	return h.ReadMetadataCall.Returns.Manifest, h.ReadMetadataCall.Returns.Image, h.ReadMetadataCall.Returns.Error
}

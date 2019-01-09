package container

import (
	"fmt"

	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

type Image struct {
	handler handler
}

type handler interface {
	ReadMetadata() (oci.Manifest, oci.Image, error)
}

func NewImage(handler handler) Image {
	return Image{
		handler: handler,
	}
}

func (u Image) ContainsHydratorAnnotation(path string) (bool, error) {
	manifest, _, err := u.handler.ReadMetadata()
	if err != nil {
		return false, fmt.Errorf("Hydrator handler failed to read metadata: %s", err)
	}

	if _, ok := manifest.Annotations["hydrator.layerAdded"]; ok {
		return true, nil
	}

	return false, nil
}

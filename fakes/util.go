package fakes

type Image struct {
	ContainsHydratorAnnotationCall struct {
		CallCount int
		Receives  struct {
			OCIImagePath string
		}
		Returns struct {
			Contains bool
			Error    error
		}
	}
}

func (u *Image) ContainsHydratorAnnotation(ociImagePath string) (bool, error) {
	u.ContainsHydratorAnnotationCall.CallCount++
	u.ContainsHydratorAnnotationCall.Receives.OCIImagePath = ociImagePath

	return u.ContainsHydratorAnnotationCall.Returns.Contains, u.ContainsHydratorAnnotationCall.Returns.Error
}

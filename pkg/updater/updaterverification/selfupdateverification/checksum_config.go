package selfupdateverification

// ChecksumConfig
type ChecksumConfig struct {
	Ref string
	URL string
}

func DefaultChecksumConfig() ChecksumConfig {
	return ChecksumConfig{
		Ref: VerificationChecksumRef,
	}
}

func (c ChecksumConfig) GetRef() string {
	return c.Ref
}

func (c *ChecksumConfig) WithURL(url string) *ChecksumConfig {
	c.URL = url
	return c
}

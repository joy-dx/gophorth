package selfupdateverification

const VerificationChecksumRef = "checksum"

type VerificationChecksum struct {
	Ref string
}

func NewVerificationChecksum() *VerificationChecksum {
	return &VerificationChecksum{
		Ref: VerificationChecksumRef,
	}
}

func (v *VerificationChecksum) GetRef() string {
	return v.Ref
}

func (v *VerificationChecksum) SetConfig() string {}

func (v *VerificationChecksum) Verify(artefactPath string) error {
	return nil
}

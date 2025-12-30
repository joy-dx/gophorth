package releaserdto

type ReleaseAsset struct {
	ArtefactName  string `json:"artefact_name"`
	Platform      string `json:"platform"` // e.g. "linux", "darwin"
	Arch          string `json:"arch"`     // e.g. "amd64", "arm64"
	Variant       string `json:"variant"`  // e.g. "webkit2_41", "standard"
	Version       string `json:"version"`
	DownloadURL   string `json:"download_url"`        // direct link to binary/archive
	Checksum      string `json:"checksum"`            // optional integrity hash (e.g. SHA256)
	SizeBytes     int64  `json:"size_bytes"`          // optional for display/use in updater
	Signature     string `json:"signature,omitempty"` // optional detached signature (for verification)
	SignatureType string `json:"signature_type,omitempty"`
}

func (l *ReleaseAsset) WithArch(arch string) *ReleaseAsset {
	l.Arch = arch
	return l
}

func (l *ReleaseAsset) WithArtefactName(name string) *ReleaseAsset {
	l.ArtefactName = name
	return l
}

func (l *ReleaseAsset) WithPlatform(platform string) *ReleaseAsset {
	l.Platform = platform
	return l
}

func (l *ReleaseAsset) WithVariant(variant string) *ReleaseAsset {
	l.Variant = variant
	return l
}

func (l *ReleaseAsset) WithDownloadURL(url string) *ReleaseAsset {
	l.DownloadURL = url
	return l
}

func (l *ReleaseAsset) WithChecksum(checksum string) *ReleaseAsset {
	l.Checksum = checksum
	return l
}

func (l *ReleaseAsset) WithSize(bytes int64) *ReleaseAsset {
	l.SizeBytes = bytes
	return l
}

func (l *ReleaseAsset) WithSignature(signature string) *ReleaseAsset {
	l.Signature = signature
	return l
}

func (l *ReleaseAsset) WithSignatureType(sigType string) *ReleaseAsset {
	l.SignatureType = sigType
	return l
}

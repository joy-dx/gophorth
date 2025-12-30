package cryptography

import (
	"crypto/sha1"
	"encoding/base32"
	"strings"
)

func EmailToWKD(email string) string {
	lower := strings.ToLower(strings.TrimSpace(email))
	h := sha1.Sum([]byte(lower))
	// Base32 encoding, lowercase, no padding
	enc := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(h[:])
	return strings.ToLower(enc)
}

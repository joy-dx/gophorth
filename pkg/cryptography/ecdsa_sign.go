package cryptography

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"os"
)

// ECDSASignFile creates a detached signature for a file.
// It returns a PEM-like armored signature that embeds:
// - Hash: SHA-256 of file content
// - Signature: DER-encoded ECDSA signature, Base64-encoded
//
// Note: For simplicity, we fix the hash to SHA-256. You can parameterize it
// if needed, but ensure verifier matches the same hash.
func ECDSASignFile(priv *ecdsa.PrivateKey, inputPath string) (string, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return "", fmt.Errorf("failed to read input file: %w", err)
	}

	h := sha256.Sum256(data)
	sigDER, err := ecdsaSignDER(priv, h[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign: %w", err)
	}

	// Armor: PEM-like block with Base64 signature and meta headers
	var buf bytes.Buffer
	if err := encodeECDSASignaturePEM(&buf, "ECDSA DETACHED SIGNATURE", map[string]string{
		"Hash": "SHA-256",
	}, sigDER); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ecdsaSignDER produces DER-encoded ECDSA signature (ASN.1 sequence of r, s).
func ecdsaSignDER(priv *ecdsa.PrivateKey, digest []byte) ([]byte, error) {
	r, s, err := ecdsa.Sign(rand.Reader, priv, digest)
	if err != nil {
		return nil, err
	}
	return asn1MarshalECDSASignature(r, s)
}

package cryptography

import (
	"bytes"
	"fmt"

	"github.com/ProtonMail/go-crypto/openpgp"
)

// LoadArmoredKeyRing reads an ASCII-armored key (public or private)
// and returns an openpgp.EntityList.
func LoadArmoredKeyRing(armoredData string) (openpgp.EntityList, error) {
	entityList, err := openpgp.ReadArmoredKeyRing(bytes.NewBufferString(armoredData))
	if err != nil {
		return nil, fmt.Errorf("failed to read armored keyring: %w", err)
	}
	return entityList, nil
}

// LoadBinaryKeyRing reads binary (non-armored) OpenPGP keys.
func LoadBinaryKeyRing(binaryData []byte) (openpgp.EntityList, error) {
	entityList, err := openpgp.ReadKeyRing(bytes.NewReader(binaryData))
	if err != nil {
		return nil, fmt.Errorf("failed to read binary keyring: %w", err)
	}
	return entityList, nil
}

// LoadKeyRingAuto detects armored or binary PGP key material and loads it.
func LoadKeyRingAuto(data []byte) (openpgp.EntityList, error) {
	trimmed := bytes.TrimSpace(data)

	if bytes.HasPrefix(trimmed, []byte("-----BEGIN PGP")) {
		// ASCII-armored key
		entityList, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(trimmed))
		if err != nil {
			return nil, fmt.Errorf("failed to read armored key: %w", err)
		}
		return entityList, nil
	}

	// Default to binary format
	entityList, err := openpgp.ReadKeyRing(bytes.NewReader(trimmed))
	if err != nil {
		return nil, fmt.Errorf("failed to read binary key: %w", err)
	}
	return entityList, nil
}

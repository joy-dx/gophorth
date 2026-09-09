package cryptography

import (
	"bytes"
	"fmt"
	"os"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
)

// PGPVerifyFile verifies a detached signature against a given file
// using a public key. Returns nil if valid, error if not.
func PGPVerifyFile(entityList openpgp.EntityList, inputPath string, signature bytes.Buffer) error {

	inFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inFile.Close()

	// Verify the signature
	_, err = openpgp.CheckArmoredDetachedSignature(
		entityList, // keyring
		inFile,     // signed data
		&signature, // signature
		&packet.Config{},
	)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

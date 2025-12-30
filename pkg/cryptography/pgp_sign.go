package cryptography

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/ProtonMail/go-crypto/openpgp"
)

// PGPSignFile creates a detached ASCII-armored signature for a file.
// It outputs a .asc (detached) signature, similar to "pgp --detach-sign".
func PGPSignFile(entityList openpgp.EntityList, inputPath string) (string, error) {

	if len(entityList) == 0 {
		return "", errors.New("no usable keys found in private key file")
	}
	entity := entityList[0]

	// Open target file
	in, err := os.Open(inputPath)
	if err != nil {
		return "", fmt.Errorf("failed to open input file: %w", err)
	}
	defer in.Close()

	// Create detached (ASCII-armored) signature
	var signatureBuffer bytes.Buffer
	if err := openpgp.ArmoredDetachSign(&signatureBuffer, entity, in, nil); err != nil {
		return "", fmt.Errorf("failed to sign file: %w", err)
	}

	return signatureBuffer.String(), nil
}

package cryptography

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"strings"
)

func InterfaceToChecksum(inputInterface interface{}) (string, error) {
	rawMessage, err := json.Marshal(inputInterface)
	if err != nil {
		return "", fmt.Errorf("error encoding struct to JSON: %s", err)
	}
	// Calculate CRC32 checksum from the serialized bytes
	checksum := crc32.ChecksumIEEE(rawMessage)
	return fmt.Sprintf("%08x", checksum), nil
}

// sha256SumFile computes the SHA-256 checksum for a given file path.
// It returns the checksum as a hex-encoded string.
func Sha256SumFile(path string) (string, error) {

	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to hash file: %w", err)
	}

	sum := hash.Sum(nil)
	return hex.EncodeToString(sum), nil
}

// sha256SumBatch computes SHA-256 checksums for multiple files and
// returns a string in the typical "checksums.txt" format.
func Sha256SumBatch(paths []string) (string, error) {
	var sb strings.Builder

	for _, path := range paths {
		sum, err := Sha256SumFile(path)
		if err != nil {
			return "", err
		}
		// sha256sum format: "<checksum>  <filename>"
		sb.WriteString(fmt.Sprintf("%s  %s\n", sum, path))
	}

	return sb.String(), nil
}

func Sha256SumVerify(path string, checksum string) error {
	targetHash, hashErr := Sha256SumFile(path)
	if hashErr != nil {
		return hashErr
	}

	if checksum != targetHash {
		return errors.New("invalid checksum")
	}
	return nil
}

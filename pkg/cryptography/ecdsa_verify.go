package cryptography

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
)

// ECDSAVerifyFile verifies a detached ECDSA signature against the file.
func ECDSAVerifyFile(pub *ecdsa.PublicKey, inputPath string, armoredSignature string) error {
	block, headers, sigDER, err := decodeECDSASignaturePEM(armoredSignature)
	if err != nil {
		return err
	}
	if block != "ECDSA DETACHED SIGNATURE" {
		return fmt.Errorf("unexpected signature block: %s", block)
	}
	if hName := headers["Hash"]; hName != "SHA-256" {
		return fmt.Errorf("unsupported or mismatched hash: %s", hName)
	}

	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}
	h := sha256.Sum256(data)

	ok, err := ecdsaVerifyDER(pub, h[:], sigDER)
	if err != nil {
		return fmt.Errorf("verification error: %w", err)
	}
	if !ok {
		return errors.New("signature verification failed")
	}
	return nil
}

// ecdsaVerifyDER verifies DER-encoded ECDSA signature.
func ecdsaVerifyDER(pub *ecdsa.PublicKey, digest []byte, sigDER []byte) (bool, error) {
	r, s, err := asn1UnmarshalECDSASignature(sigDER)
	if err != nil {
		return false, err
	}
	return ecdsa.Verify(pub, digest, r, s), nil
}

type ecdsaSignature struct {
	R, S *big.Int
}

func asn1MarshalECDSASignature(r, s *big.Int) ([]byte, error) {
	return asn1.Marshal(ecdsaSignature{R: r, S: s})
}
func asn1UnmarshalECDSASignature(der []byte) (*big.Int, *big.Int, error) {
	var sig ecdsaSignature
	if _, err := asn1.Unmarshal(der, &sig); err != nil {
		return nil, nil, err
	}
	if sig.R == nil || sig.S == nil {
		return nil, nil, errors.New("invalid ECDSA signature")
	}
	return sig.R, sig.S, nil
}

// Simple PEM-like armoring for the detached signature blob.
// We embed simple headers followed by Base64 of DER signature.
func encodeECDSASignaturePEM(w io.Writer, typ string, headers map[string]string, der []byte) error {
	if _, err := fmt.Fprintf(w, "-----BEGIN %s-----\n", typ); err != nil {
		return err
	}
	for k, v := range headers {
		if _, err := fmt.Fprintf(w, "%s: %s\n", k, v); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, "\n"); err != nil {
		return err
	}
	enc := base64.StdEncoding.EncodeToString(der)
	// Wrap at 64 chars
	const line = 64
	for i := 0; i < len(enc); i += line {
		j := i + line
		if j > len(enc) {
			j = len(enc)
		}
		if _, err := io.WriteString(w, enc[i:j]+"\n"); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w, "-----END %s-----\n", typ)
	return err
}

func decodeECDSASignaturePEM(armored string) (blockType string, headers map[string]string, der []byte, err error) {
	headers = map[string]string{}
	lines := bytes.Split([]byte(armored), []byte("\n"))

	if len(lines) < 3 {
		return "", nil, nil, errors.New("invalid signature armor")
	}

	if !bytes.HasPrefix(lines[0], []byte("-----BEGIN ")) || !bytes.HasSuffix(lines[0], []byte("-----")) {
		return "", nil, nil, errors.New("invalid armor header")
	}
	blockType = string(bytes.TrimSuffix(bytes.TrimPrefix(lines[0], []byte("-----BEGIN ")), []byte("-----")))

	i := 1
	for ; i < len(lines); i++ {
		if len(lines[i]) == 0 {
			i++
			break
		}
		parts := bytes.SplitN(lines[i], []byte(":"), 2)
		if len(parts) != 2 {
			return "", nil, nil, errors.New("invalid header line")
		}
		k := string(bytes.TrimSpace(parts[0]))
		v := string(bytes.TrimSpace(parts[1]))
		headers[k] = v
	}

	var b64Buf bytes.Buffer
	for ; i < len(lines); i++ {
		if bytes.HasPrefix(lines[i], []byte("-----END ")) {
			break
		}
		b64Buf.Write(lines[i])
	}
	raw, decErr := base64.StdEncoding.DecodeString(b64Buf.String())
	if decErr != nil {
		return "", nil, nil, fmt.Errorf("base64 decode failed: %w", decErr)
	}
	return blockType, headers, raw, nil
}

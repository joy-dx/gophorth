package cryptography

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

type ECDSACurve string

const (
	ECDSACurveP256 ECDSACurve = "P-256"
	ECDSACurveP384 ECDSACurve = "P-384"
	ECDSACurveP521 ECDSACurve = "P-521"
)

type ECDSACreateKeyCfg struct {
	// Metadata fields included for parity with PGP config; they are not embedded
	// into the ECDSA keys themselves (no native user ID in raw ECDSA keys).
	Comment string
	Email   string
	Name    string

	// Curve selection
	Curve ECDSACurve
}

func DefaultECDSACreateKeyCfg() *ECDSACreateKeyCfg {
	return &ECDSACreateKeyCfg{
		Comment: "JoyDX managed ECDSA key",
		Curve:   ECDSACurveP256,
	}
}

func (c *ECDSACreateKeyCfg) WithEmail(email string) *ECDSACreateKeyCfg {
	c.Email = email
	return c
}
func (c *ECDSACreateKeyCfg) WithComment(comment string) *ECDSACreateKeyCfg {
	c.Comment = comment
	return c
}
func (c *ECDSACreateKeyCfg) WithName(name string) *ECDSACreateKeyCfg {
	c.Name = name
	return c
}
func (c *ECDSACreateKeyCfg) WithCurve(curve ECDSACurve) *ECDSACreateKeyCfg {
	c.Curve = curve
	return c
}

func curveFromCfg(c ECDSACurve) (elliptic.Curve, error) {
	switch c {
	case ECDSACurveP256:
		return elliptic.P256(), nil
	case ECDSACurveP384:
		return elliptic.P384(), nil
	case ECDSACurveP521:
		return elliptic.P521(), nil
	default:
		return nil, fmt.Errorf("unsupported curve: %s", c)
	}
}

// ECDSACreateKey generates an ECDSA key pair and returns
// - privPEM: PEM-encoded EC PRIVATE KEY (SEC1 / ASN.1)
// - pubPEM: PEM-encoded PUBLIC KEY (SubjectPublicKeyInfo / X.509)
// This keeps them interoperable with common tooling (OpenSSL, Go x509, etc.).
func ECDSACreateKey(cfg ECDSACreateKeyCfg) (privPEM string, pubPEM string, err error) {
	curve, err := curveFromCfg(cfg.Curve)
	if err != nil {
		return "", "", err
	}

	priv, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("ecdsa key generation failed: %w", err)
	}

	// Private key in SEC1 (ASN.1, DER)
	derPriv, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return "", "", fmt.Errorf("marshal EC private key failed: %w", err)
	}
	privBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: derPriv,
	}
	var privBuf bytes.Buffer
	if err := pem.Encode(&privBuf, privBlock); err != nil {
		return "", "", fmt.Errorf("pem encode private key failed: %w", err)
	}

	// Public key in SPKI (X.509 SubjectPublicKeyInfo)
	derPub, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("marshal public key failed: %w", err)
	}
	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPub,
	}
	var pubBuf bytes.Buffer
	if err := pem.Encode(&pubBuf, pubBlock); err != nil {
		return "", "", fmt.Errorf("pem encode public key failed: %w", err)
	}

	return privBuf.String(), pubBuf.String(), nil
}

// ParseECDSAPrivateKeyFromPEM parses a PEM-encoded "EC PRIVATE KEY".
func ParseECDSAPrivateKeyFromPEM(pemStr string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil || block.Type != "EC PRIVATE KEY" {
		return nil, errors.New("invalid EC private key PEM")
	}
	priv, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse EC private key failed: %w", err)
	}
	return priv, nil
}

// ParseECDSAPublicKeyFromPEM parses a PEM-encoded "PUBLIC KEY" (SPKI).
func ParseECDSAPublicKeyFromPEM(pemStr string) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("invalid public key PEM")
	}
	pubAny, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key failed: %w", err)
	}
	pub, ok := pubAny.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not ECDSA")
	}
	return pub, nil
}

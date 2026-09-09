package cryptography

import (
	"bytes"
	"crypto"
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	pmopenpgp "github.com/ProtonMail/go-crypto/openpgp"
	pmarmor "github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
)

type KeyInfo struct {
	Format    string // "PGP", "X509", or "SSH"
	Kind      string // "Public" or "Private"
	Algorithm string // "RSA", "ECDSA", "Ed25519", etc.
	Detail    string // curve name, bits, encrypted, etc.
}

func DetectSignatureInformation(data []byte) (*KeyInfo, error) {
	// Try PGP armor first: private, public
	s := strings.TrimSpace(string(data))
	if strings.HasPrefix(s, "-----BEGIN PGP PRIVATE KEY BLOCK-----") ||
		strings.HasPrefix(s, "-----BEGIN PGP SECRET KEY BLOCK-----") {
		return detectPGPPrivate(data)
	}
	if strings.HasPrefix(s, "-----BEGIN PGP PUBLIC KEY BLOCK-----") {
		return detectPGPPublic(data)
	}
	if strings.HasPrefix(s, "-----BEGIN PGP SIGNATURE-----") {
		return detectPGPSignature(data)
	}

	// OpenSSH formats
	if strings.HasPrefix(s, "ssh-") {
		return detectSSHPublic(s)
	}
	if strings.HasPrefix(s, "-----BEGIN OPENSSH PRIVATE KEY-----") {
		return &KeyInfo{
			Format:    "SSH",
			Kind:      "Private",
			Algorithm: "Unknown",
			Detail:    "OpenSSH private key (not X.509/PGP)",
		}, nil
	}

	// PEM: iterate all blocks and pick the best
	if strings.HasPrefix(s, "-----BEGIN ") {
		return detectPEMMulti(data)
	}

	return nil, errors.New("unknown key format")
}

func detectPGPSignature(data []byte) (*KeyInfo, error) {
	block, err := pmarmor.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	// ProtonMail’s armor types typically include "PGP SIGNATURE"
	// Reference: pmopenpgp.SignatureType
	if !strings.EqualFold(block.Type, "PGP SIGNATURE") && block.Type != "SIGNATURE" {
		return nil, fmt.Errorf("unexpected PGP block type: %s", block.Type)
	}

	// Parse packets inside the armored signature
	var rdr = packet.NewReader(block.Body)
	for {
		pkt, err := rdr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		switch sig := pkt.(type) {
		case *packet.Signature:
			algo := mapPGPPubKeyAlgo(sig.PubKeyAlgo)
			detail := mapPGPHashAlgo(sig.Hash)
			issuer := issuerString(sig)
			if issuer != "" {
				if detail != "" {
					detail += ", "
				}
				detail += issuer
			}
			return &KeyInfo{
				Format:    "PGP",
				Kind:      "Signature",
				Algorithm: algo,
				Detail:    detail,
			}, nil
		case *packet.OnePassSignature:
			// One-pass can appear in inline signatures; still identify algo/hash
			algo := mapPGPPubKeyAlgo(sig.PubKeyAlgo)
			detail := mapPGPHashAlgo(sig.Hash)
			return &KeyInfo{
				Format:    "PGP",
				Kind:      "Signature",
				Algorithm: algo,
				Detail:    detail + " (one-pass)",
			}, nil
		default:
			// Skip unrelated packets if any
		}
	}

	// If we got here, it was a signature block but no signature packet parsed.
	return &KeyInfo{
		Format:    "PGP",
		Kind:      "Signature",
		Algorithm: "Unknown",
		Detail:    "no signature packet found",
	}, nil
}

// Helpers

func mapPGPPubKeyAlgo(a packet.PublicKeyAlgorithm) string {
	switch a {
	case packet.PubKeyAlgoRSA, packet.PubKeyAlgoRSAEncryptOnly, packet.PubKeyAlgoRSASignOnly:
		return "RSA"
	case packet.PubKeyAlgoDSA:
		return "DSA"
	case packet.PubKeyAlgoECDSA:
		return "ECDSA"
	case packet.PubKeyAlgoEdDSA:
		return "EdDSA"
	case packet.PubKeyAlgoElGamal:
		return "ElGamal"
	case packet.PubKeyAlgoECDH:
		return "ECDH"
	default:
		return fmt.Sprintf("Unknown(%d)", a)
	}
}

func mapPGPHashAlgo(h crypto.Hash) string {
	// packet.Signature.Hash is a crypto.Hash
	switch h {
	case crypto.SHA256:
		return "SHA-256"
	case crypto.SHA384:
		return "SHA-384"
	case crypto.SHA512:
		return "SHA-512"
	case crypto.SHA224:
		return "SHA-224"
	case crypto.SHA1:
		return "SHA-1"
	default:
		if h != 0 {
			return fmt.Sprintf("Hash(%d)", int(h))
		}
		return ""
	}
}

func issuerString(sig *packet.Signature) string {
	// Prefer issuer fingerprint if available; fallback to issuer key ID.
	if len(sig.IssuerFingerprint) > 0 {
		// hex uppercase
		return "issuer fpr " + strings.ToUpper(fmt.Sprintf("%x", sig.IssuerFingerprint))
	}
	if sig.IssuerKeyId != nil {
		return fmt.Sprintf("issuer keyid %016X", *sig.IssuerKeyId)
	}
	return ""
}

func detectPEMMulti(data []byte) (*KeyInfo, error) {
	var best *KeyInfo
	input := data
	for {
		block, rest := pem.Decode(input)
		if block == nil {
			break
		}
		info := classifyPEMBlock(block)
		if info != nil {
			if betterKey(info, best) {
				best = info
			}
		}
		input = rest
		if len(rest) == 0 {
			break
		}
	}
	if best != nil {
		return best, nil
	}
	return nil, errors.New("no recognizable key in PEM")
}

func classifyPEMBlock(block *pem.Block) *KeyInfo {
	switch block.Type {
	case "PUBLIC KEY":
		info, err := detectX509Public(block.Bytes)
		if err == nil {
			return info
		}
	case "PRIVATE KEY":
		if info, _ := detectX509PrivatePKCS8(block.Bytes); info != nil {
			return info
		}
	case "RSA PRIVATE KEY":
		if info, err := detectX509PrivateRSA(block.Bytes); err == nil {
			return info
		}
	case "EC PRIVATE KEY":
		if info, err := detectX509PrivateEC(block.Bytes); err == nil {
			return info
		}
	case "ENCRYPTED PRIVATE KEY":
		return &KeyInfo{
			Format:    "X509",
			Kind:      "Private",
			Algorithm: "Unknown",
			Detail:    "encrypted PKCS#8",
		}
	default:
		// Try generic hints
		if strings.Contains(block.Type, "PUBLIC KEY") {
			if info, err := detectX509Public(block.Bytes); err == nil {
				return info
			}
		}
		if strings.Contains(block.Type, "PRIVATE KEY") {
			// Unknown private type; still report private
			return &KeyInfo{
				Format:    "X509",
				Kind:      "Private",
				Algorithm: "Unknown",
				Detail:    fmt.Sprintf("unrecognized PEM type: %s", block.Type),
			}
		}
	}
	return nil
}

func betterKey(candidate, current *KeyInfo) bool {
	if current == nil {
		return true
	}
	// Prefer Private over Public
	if candidate.Kind != current.Kind {
		return candidate.Kind == "Private"
	}
	// Prefer known algorithm over Unknown
	if current.Algorithm == "Unknown" && candidate.Algorithm != "Unknown" {
		return true
	}
	// Otherwise keep current
	return false
}

func detectX509Public(der []byte) (*KeyInfo, error) {
	pub, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, err
	}
	switch k := pub.(type) {
	case *rsa.PublicKey:
		return &KeyInfo{"X509", "Public", "RSA", fmt.Sprintf("%d bits", k.N.BitLen())}, nil
	case *ecdsa.PublicKey:
		return &KeyInfo{"X509", "Public", "ECDSA", curveName(k.Curve)}, nil
	case ed25519.PublicKey:
		return &KeyInfo{"X509", "Public", "Ed25519", ""}, nil
	default:
		return nil, errors.New("unsupported X509 public key type")
	}
}

func detectX509PrivatePKCS8(der []byte) (*KeyInfo, error) {
	priv, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return &KeyInfo{
			Format:    "X509",
			Kind:      "Private",
			Algorithm: "Unknown",
			Detail:    "unparseable PKCS#8 (possibly encrypted or unsupported)",
		}, nil
	}
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &KeyInfo{"X509", "Private", "RSA", fmt.Sprintf("%d bits", k.N.BitLen())}, nil
	case *ecdsa.PrivateKey:
		return &KeyInfo{"X509", "Private", "ECDSA", curveName(k.Curve)}, nil
	case ed25519.PrivateKey:
		return &KeyInfo{"X509", "Private", "Ed25519", ""}, nil
	default:
		return &KeyInfo{"X509", "Private", "Unknown", "unsupported PKCS#8 key type"}, nil
	}
}

func detectX509PrivateRSA(der []byte) (*KeyInfo, error) {
	k, err := x509.ParsePKCS1PrivateKey(der)
	if err != nil {
		return nil, err
	}
	return &KeyInfo{"X509", "Private", "RSA", fmt.Sprintf("%d bits", k.N.BitLen())}, nil
}

func detectX509PrivateEC(der []byte) (*KeyInfo, error) {
	k, err := x509.ParseECPrivateKey(der)
	if err != nil {
		return nil, err
	}
	return &KeyInfo{"X509", "Private", "ECDSA", curveName(k.Curve)}, nil
}

func curveName(c elliptic.Curve) string {
	switch c {
	case elliptic.P224():
		return "P-224"
	case elliptic.P256():
		return "P-256"
	case elliptic.P384():
		return "P-384"
	case elliptic.P521():
		return "P-521"
	default:
		if c != nil && c.Params() != nil && c.Params().Name != "" {
			return c.Params().Name
		}
		return "unknown"
	}
}

func detectPGPPublic(data []byte) (*KeyInfo, error) {
	block, err := pmarmor.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if block.Type != pmopenpgp.PublicKeyType {
		return nil, fmt.Errorf("unexpected PGP block type: %s", block.Type)
	}
	entities, err := pmopenpgp.ReadArmoredKeyRing(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if len(entities) == 0 || entities[0].PrimaryKey == nil {
		return nil, errors.New("no PGP primary key found")
	}
	pk := entities[0].PrimaryKey
	algo, detail := mapPGPAlgoAndCurve(pk)
	return &KeyInfo{"PGP", "Public", algo, detail}, nil
}

func detectPGPPrivate(data []byte) (*KeyInfo, error) {
	block, err := pmarmor.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if block.Type != pmopenpgp.PrivateKeyType && block.Type != "PGP SECRET KEY BLOCK" {
		return nil, fmt.Errorf("unexpected PGP block type: %s", block.Type)
	}

	entities, err := pmopenpgp.ReadArmoredKeyRing(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if len(entities) == 0 || entities[0].PrimaryKey == nil {
		return nil, errors.New("no PGP primary key found")
	}

	ent := entities[0]
	if ent.PrivateKey != nil {
		algo, detail := mapPGPAlgoAndCurve(&ent.PrivateKey.PublicKey)

		// Add size details when available by inspecting concrete private key type.
		switch sk := ent.PrivateKey.PrivateKey.(type) {
		case *rsa.PrivateKey:
			if sk.N != nil {
				if detail != "" {
					detail += ", "
				}
				detail += fmt.Sprintf("%d bits", sk.N.BitLen())
			}
		case *dsa.PrivateKey:
			if sk.Parameters.P != nil {
				if detail != "" {
					detail += ", "
				}
				detail += fmt.Sprintf("%d bits", sk.Parameters.P.BitLen())
			}
			// For ECDSA/EdDSA/X25519/X448 we already expose curve name via mapPGPAlgoAndCurve.
		default:
			// no size detail
		}
		if ent.PrivateKey.Encrypted {
			if detail != "" {
				detail += ", "
			}
			detail += "encrypted"
		}
		return &KeyInfo{"PGP", "Private", algo, detail}, nil
	}

	// Fallback to public params
	algo, detail := mapPGPAlgoAndCurve(ent.PrimaryKey)
	return &KeyInfo{"PGP", "Private", algo, detail + " (no private material parsed)"}, nil
}

// mapPGPAlgoAndCurve extracts algorithm and attempts to name ECC curves via OIDs.
func mapPGPAlgoAndCurve(pk *packet.PublicKey) (algo, detail string) {
	switch pk.PubKeyAlgo {
	case packet.PubKeyAlgoRSA, packet.PubKeyAlgoRSAEncryptOnly, packet.PubKeyAlgoRSASignOnly:
		bitLength, bitLengthErr := pk.BitLength()
		if bitLengthErr != nil {
			return "PubKeyAlgoRSASignOnly unable to determine bit length", ""
		}
		return "RSA", fmt.Sprintf("%d bits", bitLength)
	case packet.PubKeyAlgoDSA:
		return "DSA", ""
	case packet.PubKeyAlgoElGamal:
		return "ElGamal", ""
	case packet.PubKeyAlgoECDSA:
		return "ECDSA", curveFromPGPCurveOID(pk)
	case packet.PubKeyAlgoECDH:
		if cn := curveFromPGPCurveOID(pk); cn != "" {
			return "ECDH", cn
		}
		return "ECDH", ""
	case packet.PubKeyAlgoEdDSA:
		if cn := curveFromPGPCurveOID(pk); cn != "" {
			return "EdDSA", cn
		}
		// Default assumption when OID not exposed
		return "EdDSA", "Ed25519"
	default:
		return fmt.Sprintf("Unknown(%d)", pk.PubKeyAlgo), ""
	}
}

// curveFromPGPCurveOID tries to read an asn1.ObjectIdentifier named CurveOID on pk
// or on a nested Params field. Uses reflection to avoid hard coupling to ProtonMail
// internal struct evolution. Returns "" if unavailable.
func curveFromPGPCurveOID(pk *packet.PublicKey) string {
	if pk == nil {
		return ""
	}

	type oidT = asn1.ObjectIdentifier
	oidType := reflect.TypeOf(oidT{})

	readCurveOID := func(v any) (oidT, bool) {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Pointer {
			rv = rv.Elem()
		}
		if rv.Kind() != reflect.Struct {
			return nil, false
		}
		f := rv.FieldByName("CurveOID")
		if !f.IsValid() || f.Type() != oidType || !f.CanInterface() {
			return nil, false
		}
		oid := f.Interface().(oidT)
		if len(oid) == 0 {
			return nil, false
		}
		return oid, true
	}

	// Direct CurveOID on pk
	if oid, ok := readCurveOID(pk); ok {
		return oidToCurveName(oid)
	}

	// Params.CurveOID inside pk, if present
	rv := reflect.ValueOf(pk)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Struct {
		params := rv.FieldByName("Params")
		if params.IsValid() && params.CanInterface() {
			if oid, ok := readCurveOID(params.Interface()); ok {
				return oidToCurveName(oid)
			}
		}
	}

	return ""
}

func oidToCurveName(oid asn1.ObjectIdentifier) string {
	switch oid.String() {
	case "1.2.840.10045.3.1.7":
		return "P-256"
	case "1.3.132.0.34":
		return "P-384"
	case "1.3.132.0.35":
		return "P-521"
	case "1.3.132.0.33":
		return "secp224r1"
	case "1.3.132.0.10":
		return "secp256k1"
	case "1.3.101.112":
		return "Ed25519"
	case "1.3.101.113":
		return "Ed448"
	case "1.3.101.110":
		return "X25519"
	case "1.3.101.111":
		return "X448"
	default:
		return oid.String()
	}
}

// tryReflectCurveOID attempts to extract an asn1.ObjectIdentifier from
// common field layouts used by the ProtonMail fork. It returns "" on failure.
func tryReflectCurveOID(pk *packet.PublicKey) string {
	// We avoid importing reflect if you prefer; but it's small and avoids tight coupling.
	// If you dislike reflection, remove this function and depend solely on ASN.1 fallback.
	type oidLike = asn1.ObjectIdentifier

	// Small local helper to get a field by name and type
	getOID := func(v any, field string) (oidLike, bool) {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Pointer {
			rv = rv.Elem()
		}
		if rv.Kind() != reflect.Struct {
			return nil, false
		}
		f := rv.FieldByName(field)
		if !f.IsValid() {
			return nil, false
		}
		if f.Type() != reflect.TypeOf(oidLike{}) {
			return nil, false
		}
		oid := f.Interface().(oidLike)
		if len(oid) == 0 {
			return nil, false
		}
		return oid, true
	}

	// Try pk.CurveOID
	if oid, ok := getOID(pk, "CurveOID"); ok {
		return oidToCurveName(oid)
	}
	// Try pk.Params.CurveOID
	// pk may have a field Params with type containing CurveOID
	rv := reflect.ValueOf(pk)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Struct {
		params := rv.FieldByName("Params")
		if params.IsValid() {
			if oid, ok := getOID(params.Interface(), "CurveOID"); ok {
				return oidToCurveName(oid)
			}
		}
	}
	return ""
}

// curveFromPGPByCurveOID attempts to read an ASN.1 Object Identifier field named
// "CurveOID" either directly on pk or within a nested "Params" struct.
// This is best-effort and safe across ProtonMail versions; if not found, returns "".
func curveFromPGPByCurveOID(pk *packet.PublicKey) string {
	if pk == nil {
		return ""
	}

	type oid = asn1.ObjectIdentifier

	// Helper to extract a field named CurveOID with type asn1.ObjectIdentifier.
	getCurveOID := func(v any) (oid, bool) {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Pointer {
			rv = rv.Elem()
		}
		if rv.Kind() != reflect.Struct {
			return nil, false
		}
		f := rv.FieldByName("CurveOID")
		if !f.IsValid() || !f.CanInterface() {
			return nil, false
		}
		if f.Type() != reflect.TypeOf(oid{}) {
			return nil, false
		}
		got := f.Interface().(oid)
		if len(got) == 0 {
			return nil, false
		}
		return got, true
	}

	// 1) Direct field on pk
	if o, ok := getCurveOID(pk); ok {
		return oidToCurveName(o)
	}

	// 2) Nested Params.CurveOID, if present
	rv := reflect.ValueOf(pk)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Struct {
		params := rv.FieldByName("Params")
		if params.IsValid() && params.CanInterface() {
			if o, ok := getCurveOID(params.Interface()); ok {
				return oidToCurveName(o)
			}
		}
	}

	// Not found
	return ""
}

func detectSSHPublic(s string) (*KeyInfo, error) {
	// ssh-ed25519, ssh-rsa, ecdsa-sha2-nistp256, etc.
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return nil, errors.New("invalid SSH public key")
	}
	algo := fields[0]
	switch {
	case algo == "ssh-ed25519":
		return &KeyInfo{"SSH", "Public", "Ed25519", ""}, nil
	case strings.HasPrefix(algo, "ecdsa-sha2-"):
		// ecdsa-sha2-nistp256|nistp384|nistp521
		curve := strings.TrimPrefix(algo, "ecdsa-sha2-")
		switch curve {
		case "nistp256":
			return &KeyInfo{"SSH", "Public", "ECDSA", "P-256"}, nil
		case "nistp384":
			return &KeyInfo{"SSH", "Public", "ECDSA", "P-384"}, nil
		case "nistp521":
			return &KeyInfo{"SSH", "Public", "ECDSA", "P-521"}, nil
		default:
			return &KeyInfo{"SSH", "Public", "ECDSA", curve}, nil
		}
	case algo == "ssh-rsa":
		return &KeyInfo{"SSH", "Public", "RSA", ""}, nil
	default:
		return &KeyInfo{"SSH", "Public", "Unknown", algo}, nil
	}
}

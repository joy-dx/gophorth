package cryptography

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test input generator signature
type genFunc func(t *testing.T) []byte

// Helpers to generate DER/PEM for X.509 cases
func pemBlock(t *testing.T, typ string, der []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := pem.Encode(&buf, &pem.Block{Type: typ, Bytes: der}); err != nil {
		t.Fatalf("pem encode: %v", err)
	}
	return buf.Bytes()
}

func genRSAPrivatePKCS1(bits int) genFunc {
	return func(t *testing.T) []byte {
		k, err := rsa.GenerateKey(rand.Reader, bits)
		if err != nil {
			t.Fatalf("rsa gen: %v", err)
		}
		return pemBlock(t, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(k))
	}
}

func genECDSAPrivateSEC1(curve elliptic.Curve) genFunc {
	return func(t *testing.T) []byte {
		k, err := ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			t.Fatalf("ecdsa gen: %v", err)
		}
		der, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			t.Fatalf("ec marshal: %v", err)
		}
		return pemBlock(t, "EC PRIVATE KEY", der)
	}
}

func genPKCS8Private(kind string) genFunc {
	return func(t *testing.T) []byte {
		switch kind {
		case "rsa":
			k, _ := rsa.GenerateKey(rand.Reader, 2048)
			der, _ := x509.MarshalPKCS8PrivateKey(k)
			return pemBlock(t, "PRIVATE KEY", der)
		case "ecdsa":
			k, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			der, _ := x509.MarshalPKCS8PrivateKey(k)
			return pemBlock(t, "PRIVATE KEY", der)
		case "ed25519":
			_, sk, _ := ed25519.GenerateKey(rand.Reader)
			der, _ := x509.MarshalPKCS8PrivateKey(sk)
			return pemBlock(t, "PRIVATE KEY", der)
		default:
			t.Fatalf("unsupported pkcs8 kind %s", kind)
			return nil
		}
	}
}

func genPKIXPublic(kind string) genFunc {
	return func(t *testing.T) []byte {
		var pub any
		switch kind {
		case "rsa":
			k, _ := rsa.GenerateKey(rand.Reader, 1024)
			pub = &k.PublicKey
		case "p256":
			k, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			pub = &k.PublicKey
		case "ed25519":
			_, sk, _ := ed25519.GenerateKey(rand.Reader)
			pub = sk.Public().(ed25519.PublicKey)
		default:
			t.Fatalf("unsupported pkix kind %s", kind)
		}
		der, err := x509.MarshalPKIXPublicKey(pub)
		if err != nil {
			t.Fatalf("pkix marshal: %v", err)
		}
		return pemBlock(t, "PUBLIC KEY", der)
	}
}

func genConcat(a, b genFunc) genFunc {
	return func(t *testing.T) []byte {
		return append(a(t), b(t)...)
	}
}

func loadTestdata(t *testing.T, filename string) []byte {
	t.Helper()
	p := filepath.Join("testdata", filename)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Skipf("missing fixture %s: %v", p, err)
	}
	return b
}

type wantKey struct {
	format    string
	kind      string
	algorithm string
	// detailMatch is a substring that must be present in Detail (if non-empty)
	// You can use it to assert bits or curve names loosely.
	detailMatch string
}

// Golden table
func TestDetectSignatureInformation_Golden(t *testing.T) {
	tests := []struct {
		name string
		// One of data or gen must be provided
		data []byte
		gen  genFunc
		// For optional fixtures (PGP), set skipIfEmpty to true to skip when data is empty
		skipIfEmpty bool
		want        wantKey
	}{
		// X.509 Private
		{
			name: "x509_rsa_private_pkcs1",
			gen:  genRSAPrivatePKCS1(1024),
			want: wantKey{"X509", "Private", "RSA", "bits"},
		},
		{
			name: "x509_ecdsa_private_sec1_p256",
			gen:  genECDSAPrivateSEC1(elliptic.P256()),
			want: wantKey{"X509", "Private", "ECDSA", "P-256"},
		},
		{
			name: "x509_private_pkcs8_ed25519",
			gen:  genPKCS8Private("ed25519"),
			want: wantKey{"X509", "Private", "Ed25519", ""},
		},
		{
			name: "x509_private_pkcs8_rsa",
			gen:  genPKCS8Private("rsa"),
			want: wantKey{"X509", "Private", "RSA", "bits"},
		},
		{
			name: "x509_private_pkcs8_ecdsa",
			gen:  genPKCS8Private("ecdsa"),
			want: wantKey{"X509", "Private", "ECDSA", "P-256"},
		},
		{
			name: "x509_encrypted_pkcs8_marker",
			data: func() []byte {
				var buf bytes.Buffer
				_ = pem.Encode(&buf, &pem.Block{
					Type:  "ENCRYPTED PRIVATE KEY",
					Bytes: []byte{0x30, 0x82, 0x01, 0x00}, // dummy
				})
				return buf.Bytes()
			}(),
			want: wantKey{"X509", "Private", "Unknown", "encrypted"},
		},
		// X.509 Public
		{
			name: "x509_public_rsa",
			gen:  genPKIXPublic("rsa"),
			want: wantKey{"X509", "Public", "RSA", "bits"},
		},
		{
			name: "x509_public_ecdsa_p256",
			gen:  genPKIXPublic("p256"),
			want: wantKey{"X509", "Public", "ECDSA", "P-256"},
		},
		{
			name: "x509_public_ed25519",
			gen:  genPKIXPublic("ed25519"),
			want: wantKey{"X509", "Public", "Ed25519", ""},
		},
		// PEM concatenation: prefer Private over Public regardless of order
		{
			name: "pem_concat_public_then_private_prefers_private",
			gen:  genConcat(genPKIXPublic("rsa"), genRSAPrivatePKCS1(1024)),
			want: wantKey{"X509", "Private", "RSA", "bits"},
		},
		{
			name: "pem_concat_private_then_public_prefers_private",
			gen:  genConcat(genRSAPrivatePKCS1(1024), genPKIXPublic("rsa")),
			want: wantKey{"X509", "Private", "RSA", "bits"},
		},
		// SSH publics
		{
			name: "ssh_public_ed25519",
			data: []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIMockKeyMaterial user@example"),
			want: wantKey{"SSH", "Public", "Ed25519", ""},
		},
		{
			name: "ssh_public_rsa",
			data: []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7 mock user@example"),
			want: wantKey{"SSH", "Public", "RSA", ""},
		},
		{
			name: "ssh_public_ecdsa_p256",
			data: []byte("ecdsa-sha2-nistp256 AAAA... user"),
			want: wantKey{"SSH", "Public", "ECDSA", "P-256"},
		},
		{
			name: "ssh_public_ecdsa_p384",
			data: []byte("ecdsa-sha2-nistp384 AAAA... user"),
			want: wantKey{"SSH", "Public", "ECDSA", "P-384"},
		},
		{
			name: "ssh_public_ecdsa_p521",
			data: []byte("ecdsa-sha2-nistp521 AAAA... user"),
			want: wantKey{"SSH", "Public", "ECDSA", "P-521"},
		},
		{
			name: "ssh_public_unknown",
			data: []byte("ssh-unknown AAAA... user"),
			want: wantKey{"SSH", "Public", "Unknown", "ssh-unknown"},
		},
		// Unknown PEM type should error; we test via a dedicated test below.
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var in []byte
			if tc.data != nil {
				in = tc.data
			} else if tc.gen != nil {
				in = tc.gen(t)
			}
			if tc.skipIfEmpty && len(in) == 0 {
				t.Skip("skipping empty input")
			}
			info, err := DetectSignatureInformation(in)
			if err != nil {
				t.Fatalf("DetectSignatureInformation error: %v", err)
			}
			if info.Format != tc.want.format {
				t.Fatalf("format: got %q want %q", info.Format, tc.want.format)
			}
			if info.Kind != tc.want.kind {
				t.Fatalf("kind: got %q want %q", info.Kind, tc.want.kind)
			}
			if info.Algorithm != tc.want.algorithm {
				t.Fatalf("algorithm: got %q want %q", info.Algorithm, tc.want.algorithm)
			}
			if tc.want.detailMatch != "" && !strings.Contains(info.Detail, tc.want.detailMatch) {
				t.Fatalf("detail: got %q, want to contain %q", info.Detail, tc.want.detailMatch)
			}
		})
	}
}

func TestDetectSignatureInformation_UnknownPEMType(t *testing.T) {
	// Golden: unrecognized PEM label should return error
	var buf bytes.Buffer
	_ = pem.Encode(&buf, &pem.Block{Type: "FOO BAR", Bytes: []byte{0x01, 0x02}})
	if _, err := DetectSignatureInformation(buf.Bytes()); err == nil {
		t.Fatalf("expected error for unknown PEM type")
	}
}

func TestDetectSignatureInformation_UnknownInput(t *testing.T) {
	// Golden: random text is not a key
	if _, err := DetectSignatureInformation([]byte("this is not a key")); err == nil {
		t.Fatalf("expected error for unknown input")
	}
}

// -----------------------------
// PGP golden fixtures (optional)
// -----------------------------
//
// Provide these files under ./testdata to enable the following table.
// If a file is missing, the subtest is skipped.
//
// - pgp_rsa_pub.asc
// - pgp_rsa_sec.asc
// - pgp_rsa_sec_encrypted.asc
// - pgp_ecdsa_p256_pub.asc
// - pgp_ecdsa_p256_sec.asc
// - pgp_eddsa_ed25519_pub.asc
// - pgp_eddsa_ed25519_sec.asc
// - pgp_ecdh_x25519_pub.asc (optional)
// - pgp_ecdh_x25519_sec.asc (optional)

func TestDetectSignatureInformation_PGP_GoldenFixtures(t *testing.T) {
	type pgpCase struct {
		name     string
		file     string
		want     wantKey
		contains string // optional extra detail check
	}
	cases := []pgpCase{
		{
			name: "pgp_rsa_public",
			file: "pgp_rsa_pub.asc",
			want: wantKey{"PGP", "Public", "RSA", "bits"},
		},
		{
			name: "pgp_rsa_secret",
			file: "pgp_rsa_sec.asc",
			want: wantKey{"PGP", "Private", "RSA", "bits"},
		},
		{
			name: "pgp_rsa_secret_encrypted",
			file: "pgp_rsa_sec_encrypted.asc",
			want: wantKey{"PGP", "Private", "RSA", "encrypted"},
		},
		{
			name: "pgp_ecdsa_p256_public",
			file: "pgp_ecdsa_p256_pub.asc",
			// Curve detection is best-effort via reflection on CurveOID.
			// If available, expect P-256 in Detail.
			want: wantKey{"PGP", "Public", "ECDSA", ""}, // allow empty or curve
		},
		{
			name: "pgp_ecdsa_p256_secret",
			file: "pgp_ecdsa_p256_sec.asc",
			want: wantKey{"PGP", "Private", "ECDSA", ""}, // curve optional
		},
		{
			name: "pgp_eddsa_ed25519_public",
			file: "pgp_eddsa_ed25519_pub.asc",
			// Algorithm might be "EdDSA" or "Ed25519" depending on your mapping
			want: wantKey{"PGP", "Public", "EdDSA", ""}, // detail may be "Ed25519"
		},
		{
			name: "pgp_eddsa_ed25519_secret",
			file: "pgp_eddsa_ed25519_sec.asc",
			want: wantKey{"PGP", "Private", "EdDSA", ""}, // detail may be "Ed25519"
		},
		{
			name: "pgp_ecdh_x25519_public",
			file: "pgp_ecdh_x25519_pub.asc",
			want: wantKey{"PGP", "Public", "ECDH", ""}, // detail may be "X25519"
		},
		{
			name: "pgp_ecdh_x25519_secret",
			file: "pgp_ecdh_x25519_sec.asc",
			want: wantKey{"PGP", "Private", "ECDH", ""}, // detail may be "X25519"
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", tc.file))
			if err != nil {
				t.Skipf("missing fixture %s", tc.file)
			}
			info, derr := DetectSignatureInformation(data)
			if derr != nil {
				t.Fatalf("DetectSignatureInformation: %v", derr)
			}
			if info.Format != tc.want.format {
				t.Fatalf("format: got %q want %q", info.Format, tc.want.format)
			}
			if info.Kind != tc.want.kind {
				t.Fatalf("kind: got %q want %q", info.Kind, tc.want.kind)
			}
			// Accept "EdDSA" or "Ed25519" as equivalent for Ed keys
			if strings.HasPrefix(tc.name, "pgp_eddsa_") {
				if info.Algorithm != "EdDSA" && info.Algorithm != "Ed25519" {
					t.Fatalf("algorithm: got %q want EdDSA/Ed25519", info.Algorithm)
				}
			} else if info.Algorithm != tc.want.algorithm {
				t.Fatalf("algorithm: got %q want %q", info.Algorithm, tc.want.algorithm)
			}
			if tc.want.detailMatch != "" && !strings.Contains(info.Detail, tc.want.detailMatch) {
				t.Fatalf("detail: got %q want substring %q", info.Detail, tc.want.detailMatch)
			}
		})
	}
}

package cryptography

import (
	"bytes"
	"crypto"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
)

type PGPCreateKeyCfg struct {
	Comment          string
	Email            string
	Name             string
	OutputPrivateKey string
	OutputPublicKey  string
	Hash             crypto.Hash
	Cipher           packet.CipherFunction
	CompressionAlgo  packet.CompressionAlgo
	RSABits          int
}

func DefaultPGPCreateKeyCfg() *PGPCreateKeyCfg {
	return &PGPCreateKeyCfg{
		Comment:          "JoyDX managed key",
		OutputPrivateKey: "private.key",
		OutputPublicKey:  "public.key",
		Hash:             crypto.SHA256,
		Cipher:           packet.CipherAES256,
		CompressionAlgo:  packet.CompressionZLIB,
		RSABits:          2048,
	}
}
func (c *PGPCreateKeyCfg) WithEmail(email string) *PGPCreateKeyCfg {
	c.Email = email
	return c
}
func (c *PGPCreateKeyCfg) WithComment(comment string) *PGPCreateKeyCfg {
	c.Comment = comment
	return c
}
func (c *PGPCreateKeyCfg) WithName(name string) *PGPCreateKeyCfg {
	c.Name = name
	return c
}

func PGPCreateKey(cfg PGPCreateKeyCfg) (string, string, error) {
	entity, err := openpgp.NewEntity(cfg.Name, cfg.Comment,
		cfg.Email, &packet.Config{
			DefaultHash:            cfg.Hash,
			DefaultCipher:          cfg.Cipher,
			DefaultCompressionAlgo: cfg.CompressionAlgo,
			RSABits:                cfg.RSABits,
		})
	if err != nil {
		panic(err)
	}

	// Save the private key
	var privBuf bytes.Buffer
	privArmor, err := armor.Encode(&privBuf, openpgp.PrivateKeyType, nil)
	if err != nil {
		return "", "", err
	}
	if err := entity.SerializePrivate(privArmor, nil); err != nil {
		return "", "", err
	}
	privArmor.Close()

	// Save the public key
	var pubBuf bytes.Buffer
	pubArmor, err := armor.Encode(&pubBuf, openpgp.PublicKeyType, nil)
	if err != nil {
		return "", "", err
	}
	if err := entity.Serialize(pubArmor); err != nil {
		return "", "", err
	}
	pubArmor.Close()

	return privBuf.String(), pubBuf.String(), nil
}

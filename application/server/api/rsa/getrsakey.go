package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"

	"github.com/wumansgy/goEncrypt"
)

// structure of keys
type RsaKey struct {
	PrivateKey string
	PublicKey  string
}

// generate paired keyss
func GenerateRsaKeyBase64(bits int) (RsaKey, error) {
	if bits != 1024 && bits != 2048 {
		return RsaKey{}, goEncrypt.ErrRsaBits
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return RsaKey{}, err
	}
	return RsaKey{
		PrivateKey: base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PrivateKey(privateKey)),
		PublicKey:  base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)),
	}, nil
}

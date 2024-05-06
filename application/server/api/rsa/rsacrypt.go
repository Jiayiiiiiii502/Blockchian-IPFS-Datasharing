package rsa

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"runtime"

	log "github.com/sirupsen/logrus"
)

// Operation with rsa encryption
func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetReportCaller(true)
}

// encryption
func rsaEncrypt(plainText, publicKey []byte) (cipherText []byte, err error) {
	defer func() {
		if err := recover(); err != nil {
			switch err.(type) {
			case runtime.Error:
				log.Errorf("runtime err=%v,Check that the key or text is correct", err)
			default:
				log.Errorf("error=%v,check the cipherText ", err)
			}
		}
	}()
	pub, err := x509.ParsePKCS1PublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	pubSize, plainTextSize := pub.Size(), len(plainText)

	offSet, once := 0, pubSize-11
	buffer := bytes.Buffer{}
	for offSet < plainTextSize {
		endIndex := offSet + once
		if endIndex > plainTextSize {
			endIndex = plainTextSize
		}
		bytesOnce, err := rsa.EncryptPKCS1v15(rand.Reader, pub, plainText[offSet:endIndex])
		if err != nil {
			return nil, err
		}
		buffer.Write(bytesOnce)
		offSet = endIndex
	}
	cipherText = buffer.Bytes()
	return cipherText, nil
}

// decryption
func rsaDecrypt(cipherText, privateKey []byte) (plainText []byte, err error) {
	defer func() {
		if err := recover(); err != nil {
			switch err.(type) {
			case runtime.Error:
				log.Errorf("runtime err=%v,Check that the key or text is correct", err)
			default:
				log.Errorf("error=%v,check the cipherText ", err)
			}
		}
	}()
	pri, err := x509.ParsePKCS1PrivateKey(privateKey)
	if err != nil {
		return []byte{}, err
	}
	priSize, cipherTextSize := pri.Size(), len(cipherText)
	var offSet = 0
	var buffer = bytes.Buffer{}
	for offSet < cipherTextSize {
		endIndex := offSet + priSize
		if endIndex > cipherTextSize {
			endIndex = cipherTextSize
		}
		bytesOnce, err := rsa.DecryptPKCS1v15(rand.Reader, pri, cipherText[offSet:endIndex])
		if err != nil {
			return nil, err
		}
		buffer.Write(bytesOnce)
		offSet = endIndex
	}
	plainText = buffer.Bytes()
	return plainText, nil
}

// to base64
func RsaEncryptToBase64(plainText []byte, base64PubKey string) (base64CipherText string, err error) {
	pub, err := base64.StdEncoding.DecodeString(base64PubKey)
	if err != nil {
		return "", err
	}
	cipherBytes, err := rsaEncrypt(plainText, pub)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(cipherBytes), nil
}

// to keys
func RsaDecryptByBase64(base64CipherText, base64PriKey string) (plainText []byte, err error) {
	privateBytes, err := base64.StdEncoding.DecodeString(base64PriKey)
	if err != nil {
		return nil, err
	}
	cipherTextBytes, err := base64.StdEncoding.DecodeString(base64CipherText)
	if err != nil {
		return nil, err
	}
	return rsaDecrypt(cipherTextBytes, privateBytes)
}

// find public key by private key
func SktoPub(sk string) (pk string, err error) {
	//base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PrivateKey(privateKey))
	skdecode, _ := base64.StdEncoding.DecodeString(sk)
	skorginal, err := x509.ParsePKCS1PrivateKey(skdecode)
	if err != nil {
		return "", fmt.Errorf("ipfs get file faile,err:%v", err)

	}
	pk = base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PublicKey(&skorginal.PublicKey))
	return pk, nil
}

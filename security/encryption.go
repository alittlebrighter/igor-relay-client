package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

const nonceLen = 12

func Encrypt(key, msg []byte) (encMsg []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}

	nonce, err := NewNonce()
	if err != nil {
		return
	}

	encMsg = aesgcm.Seal(nil, nonce, msg, nil)
	encMsg = append(nonce, encMsg...)
	return
}

func EncryptToString(key, msg []byte) (encMsg string, err error) {
	data, err := Encrypt(key, msg)
	if err != nil {
		return
	}
	encMsg = base64.StdEncoding.EncodeToString(data)
	return
}

func Decrypt(key, encMsg []byte) (msg []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}

	nonce := encMsg[:nonceLen]
	msg, err = aesgcm.Open(nil, nonce, encMsg[nonceLen:], nil)
	return
}

func DecryptFromString(key []byte, encMsg string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encMsg)
	if err != nil {
		return nil, err
	}
	return Decrypt(key, data)
}

func NewNonce() (nonce []byte, err error) {
	nonce = make([]byte, nonceLen)
	_, err = io.ReadFull(rand.Reader, nonce)
	return
}

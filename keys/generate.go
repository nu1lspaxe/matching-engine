package keys

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"runtime"
)

func GenerateEd25519KeyPair(filename string) error {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	bytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}

	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: bytes,
	}

	err = writeFile(filename, pem.EncodeToMemory(block), 0600)
	if err != nil {
		return err
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return err
	}

	pubBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}

	err = writeFile(filename+".pub", pem.EncodeToMemory(pubBlock), 0644)
	if err != nil {
		return err
	}

	return nil
}

func writeFile(filename string, data []byte, permission os.FileMode) error {
	if runtime.GOOS == "windows" {
		return os.WriteFile(filename, data, 0666)
	}
	return os.WriteFile(filename, data, permission)
}

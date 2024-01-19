package internal

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func generateHostKey(fn string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("error generating RSA key: %w", err)
	}

	var privateKeyBytes []byte = x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	if err := os.MkdirAll(filepath.Dir(fn), 0700); err != nil {
		return fmt.Errorf("error creating path %s: %w", filepath.Dir(fn), err)
	}

	f, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("error creating file %s: %w", fn, err)
	}
	defer f.Close()

	if err := pem.Encode(f, privateKeyBlock); err != nil {
		return fmt.Errorf("error encoding PEM file: %w", err)
	}

	return nil
}

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/keygen <secrets-directory>")
		os.Exit(2)
	}
	directory := os.Args[1]
	if err := os.MkdirAll(directory, 0o700); err != nil {
		fail(err)
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, 3072)
	if err != nil {
		fail(err)
	}
	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		fail(err)
	}
	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		fail(err)
	}
	writePEM(filepath.Join(directory, "jwt_private_key.pem"), "PRIVATE KEY", privateDER, 0o600)
	writePEM(filepath.Join(directory, "jwt_public_key.pem"), "PUBLIC KEY", publicDER, 0o644)

	visitorSecret := make([]byte, 32)
	if _, err := rand.Read(visitorSecret); err != nil {
		fail(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "visitor_secret.txt"), []byte(base64.RawURLEncoding.EncodeToString(visitorSecret)), 0o600); err != nil {
		fail(err)
	}
	fmt.Println("created JWT key pair and visitor secret")
}

func writePEM(path, blockType string, bytes []byte, mode os.FileMode) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		fail(err)
	}
	defer file.Close()
	if err := pem.Encode(file, &pem.Block{Type: blockType, Bytes: bytes}); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

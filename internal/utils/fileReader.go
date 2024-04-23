package utils

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
)

func ReadPrivateKey(filepath string) (any, error) {
	keyFile, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyFile)
	if block == nil {
		return nil, fmt.Errorf("decode private key error")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func ReadPublicKey(filepath string) (any, error) {
	keyFile, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyFile)
	if block == nil {
		return nil, fmt.Errorf("decode public key error")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func ReadVerificationEmailTemplate(path string, behavior string, verification int) (string, error) {
	// Read the HTML file
	fileContent, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	content := string(fileContent)
	content = strings.Replace(content, "{{behavior}}", behavior, -1)
	content = strings.Replace(content, "{{verification}}", fmt.Sprint(verification), -1)
	// Print the updated HTML
	return content, nil
}

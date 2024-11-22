package internal

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

func GenerateRSAkeys() (string, string) {
	// Генерация закрытого ключа RSA (2048-битного)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println("Ошибка при генерации ключа:", err)
		return "", ""
	}

	// Преобразование закрытого ключа в PEM-формат
	privateKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)

	// Извлечение открытого ключа из закрытого
	publicKey := &privateKey.PublicKey

	// Преобразование открытого ключа в PEM-формат
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		fmt.Println("Ошибка при кодировании открытого ключа:", err)
		return "", ""
	}
	publicKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: publicKeyBytes,
		},
	)

	return string(privateKeyPEM), string(publicKeyPEM)
}

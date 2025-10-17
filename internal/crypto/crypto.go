// Package crypto with RSA encryption and decryption functions
package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/Nikolay961996/metsys/models"
)

// ParseRSAPublicKeyPEM parse RSA public key from PEM file
func ParseRSAPublicKeyPEM(filename string) (*rsa.PublicKey, error) {
	pemData, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error file reading %s: %v", filename, err)
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("error decoding PEM block")
	}

	switch block.Type {
	case "RSA PUBLIC KEY":
		// PKCS1
		return x509.ParsePKCS1PublicKey(block.Bytes)
	case "PUBLIC KEY":
		// PKIX
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error parsing PKIX public key: %v", err)
		}
		rsaPub, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("this key is not an RSA public key")
		}
		return rsaPub, nil
	default:
		return nil, fmt.Errorf("unsupported type PEM block for public key: %s", block.Type)
	}
}

// ParseRSAPrivateKeyPEM parse RSA private key from PEM file
func ParseRSAPrivateKeyPEM(filename string) (*rsa.PrivateKey, error) {
	pemData, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", filename, err)
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("cannot decode PEM block")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		// PKCS1
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		// PKCS8
		p, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error parsing PKCS8 private key: %v", err)
		}
		rsaP, ok := p.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("key in PKCS8 container is not RSA private key")
		}
		return rsaP, nil
	default:
		return nil, fmt.Errorf("unsupporter type PEM block for private key: %s", block.Type)
	}
}

// EncryptMessageWithPublicKey encrypt message with public RSA key
func EncryptMessageWithPublicKey(message []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	maxBlockSize := publicKey.Size() - 11

	if len(message) <= maxBlockSize {
		models.Log.Info("Simple encryption (1 block)")
		return rsa.EncryptPKCS1v15(rand.Reader, publicKey, message)
	} else {
		models.Log.Info(fmt.Sprintf("Block encryption (%d block)", (len(message)+maxBlockSize-1)/maxBlockSize))
		return encryptWithBlockHeader(message, publicKey)
	}
}

// DecryptMessageWithPrivateKey decrypt message with private RSA key
func DecryptMessageWithPrivateKey(encryptedMessage []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	if isBlockEncryptedFormat(encryptedMessage) {
		models.Log.Info("Block decryption")
		return decryptWithBlockHeader(encryptedMessage, privateKey)
	} else {
		models.Log.Info("Simple decryption (1 block)")
		return rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedMessage)
	}
}

func encryptWithBlockHeader(message []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	keySize := publicKey.Size()
	maxBlockSize := keySize - 11
	totalBlocks := (len(message) + maxBlockSize - 1) / maxBlockSize

	header := make([]byte, 8)
	copy(header[0:4], "RSA_M")
	binary.BigEndian.PutUint32(header[4:8], uint32(totalBlocks))

	var result []byte
	result = append(result, header...)

	for i := 0; i < len(message); i += maxBlockSize {
		end := i + maxBlockSize
		if end > len(message) {
			end = len(message)
		}

		block := message[i:end]
		encryptedBlock, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, block)
		if err != nil {
			return nil, fmt.Errorf("error block encryption %d: %v", i/maxBlockSize, err)
		}

		// add block size before blocks
		blockHeader := make([]byte, 4)
		binary.BigEndian.PutUint32(blockHeader, uint32(len(encryptedBlock)))
		result = append(result, blockHeader...)
		result = append(result, encryptedBlock...)
	}

	return result, nil
}

func decryptWithBlockHeader(encryptedMessage []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	if len(encryptedMessage) < 8 || string(encryptedMessage[0:4]) != "RSA_M" {
		return nil, fmt.Errorf("error decrypting block message format")
	}

	totalBlocks := binary.BigEndian.Uint32(encryptedMessage[4:8])
	var result []byte
	position := 8

	for i := 0; i < int(totalBlocks); i++ {
		if position+4 > len(encryptedMessage) {
			return nil, fmt.Errorf("error format: need block size")
		}

		blockSize := binary.BigEndian.Uint32(encryptedMessage[position : position+4])
		position += 4

		if position+int(blockSize) > len(encryptedMessage) {
			return nil, fmt.Errorf("error format: block is so big")
		}

		block := encryptedMessage[position : position+int(blockSize)]
		decryptedBlock, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, block)
		if err != nil {
			return nil, fmt.Errorf("error decryption %d: %v", i, err)
		}

		result = append(result, decryptedBlock...)
		position += int(blockSize)
	}

	return result, nil
}

func isBlockEncryptedFormat(data []byte) bool {
	return len(data) >= 8 && string(data[0:4]) == "RSA_M"
}

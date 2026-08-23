// Package bpjs mengimplementasikan klien Web Service Antrean Online
// BPJS Kesehatan (Mobile JKN) untuk fasilitas kesehatan tingkat pertama.
//
// Autentikasi mengikuti spesifikasi BPJS:
//   - Header X-cons-id, X-timestamp (unix detik), user_key
//   - X-signature = Base64(HMAC-SHA256(consID+"&"+timestamp, secretKey))
//   - Respons terenkripsi AES-256-CBC lalu dikompresi LZString
package bpjs

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"
)

// Signature menghasilkan nilai header X-signature.
func Signature(consID, secretKey string, timestamp int64) string {
	return hmacBase64(consID+"&"+strconv.FormatInt(timestamp, 10), secretKey)
}

// hmacBase64 menghitung Base64(HMAC-SHA256(message, key)).
func hmacBase64(message, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// CurrentTimestamp mengembalikan unix timestamp dalam satuan detik.
func CurrentTimestamp() int64 {
	return time.Now().UTC().Unix()
}

// deriveKey menurunkan kunci & IV AES dari consID+secret+timestamp
// sesuai algoritma BPJS: digest sha256 hex → biner (32 byte kunci,
// 16 byte pertama sebagai IV).
func deriveKey(consID, secretKey string, timestamp int64) (key, iv []byte, err error) {
	digest := sha256.Sum256([]byte(consID + secretKey + strconv.FormatInt(timestamp, 10)))
	return digest[:], digest[:16], nil
}

// Decrypt mendekripsi field "response" dari envelope BPJS.
// Hasilnya masih terkompresi LZString; panggil Decompress pada hasilnya.
func Decrypt(consID, secretKey string, timestamp int64, encryptedBase64 string) ([]byte, error) {
	key, iv, err := deriveKey(consID, secretKey, timestamp)
	if err != nil {
		return nil, fmt.Errorf("derive key: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return nil, fmt.Errorf("decode base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes cipher: %w", err)
	}

	if len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("panjang ciphertext tidak valid: %d", len(ciphertext))
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	return pkcs7Unpad(plaintext, aes.BlockSize)
}

// Encrypt menyandi plaintext menjadi format terenkripsi BPJS
// (dipakai untuk pengujian roundtrip dan mock server).
func Encrypt(consID, secretKey string, timestamp int64, plaintext []byte) (string, error) {
	key, iv, err := deriveKey(consID, secretKey, timestamp)
	if err != nil {
		return "", fmt.Errorf("derive key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("aes cipher: %w", err)
	}

	padded := pkcs7Pad(plaintext, aes.BlockSize)
	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(padded))
	mode.CryptBlocks(ciphertext, padded)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padLen := blockSize - len(data)%blockSize
	padding := make([]byte, padLen)
	for i := range padding {
		padding[i] = byte(padLen)
	}
	return append(data, padding...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data kosong")
	}
	if len(data)%blockSize != 0 {
		return nil, fmt.Errorf("panjang data bukan kelipatan block size")
	}
	padLen := int(data[len(data)-1])
	if padLen == 0 || padLen > blockSize || padLen > len(data) {
		return nil, fmt.Errorf("padding tidak valid")
	}
	for _, b := range data[len(data)-padLen:] {
		if int(b) != padLen {
			return nil, fmt.Errorf("padding tidak valid")
		}
	}
	return data[:len(data)-padLen], nil
}

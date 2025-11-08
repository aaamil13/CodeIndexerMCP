package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"fmt"
)

// HashFile computes SHA256 hash of a file
func HashFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// HashBytes computes SHA256 hash of byte slice
func HashBytes(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}

// GenerateID generates a unique ID for a symbol based on its file, name, kind, and line number.
func GenerateID(filePath, name, kind string, line int) string {
	data := fmt.Sprintf("%s-%s-%s-%d", filePath, name, kind, line)
	return HashBytes([]byte(data))
}

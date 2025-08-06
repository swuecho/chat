package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/rotisserie/eris"
	"golang.org/x/crypto/pbkdf2"
)

const (
	iterations = 260000
	saltSize   = 16
	keySize    = 32
)

func generateSalt() ([]byte, error) {
	salt := make([]byte, saltSize)
	_, err := rand.Read(salt)
	return salt, err
}

func GeneratePasswordHash(password string) (string, error) {
	salt, err := generateSalt()
	if err != nil {
		return "", eris.Wrap(err, "error generating salt: ")
	}

	hash := pbkdf2.Key([]byte(password), salt, iterations, keySize, sha256.New)
	encodedHash := base64.StdEncoding.EncodeToString(hash)
	encodedSalt := base64.StdEncoding.EncodeToString(salt)

	passwordHash := fmt.Sprintf("pbkdf2_sha256$%d$%s$%s", iterations, encodedSalt, encodedHash)

	return passwordHash, nil
}

func ValidatePassword(password, hash string) bool {
	fields := strings.Split(hash, "$")
	if len(fields) != 4 || fields[0] != "pbkdf2_sha256" || fields[1] != fmt.Sprintf("%d", iterations) {
		return false
	}
	encodedSalt := fields[2]
	decodedSalt, err := base64.StdEncoding.DecodeString(encodedSalt)
	if err != nil {
		return false
	}

	encodedHash := fields[3]
	decodedHash, err := base64.StdEncoding.DecodeString(encodedHash)
	if err != nil {
		return false
	}
	computedHash := pbkdf2.Key([]byte(password), decodedSalt, iterations, keySize, sha256.New)
	return subtle.ConstantTimeCompare(decodedHash, computedHash) == 1
}

func GenerateRandomPassword() (string, error) {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	password := make([]byte, 12)
	_, err := rand.Read(password)
	if err != nil {
		return "", eris.Wrap(err, "failed to generate random password")
	}
	for i := 0; i < len(password); i++ {
		password[i] = letters[int(password[i])%len(letters)]
	}
	return string(password), nil
}

package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

// GenerateOTP generates a random 6-digit OTP code as a string.
func GenerateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n), nil
}

// HashOTP hashes the 6-digit OTP using bcrypt for safe database storage.
func HashOTP(code string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(code), 10)
	return string(bytes), err
}

// CheckOTP verifies the provided code against the hashed OTP.
func CheckOTP(code, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(code))
	return err == nil
}

// GenerateToken generates a secure 32-byte random hex string.
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashToken hashes a session token using SHA-256 for fast lookup.
// We use SHA-256 here instead of bcrypt because tokens are high entropy.
func HashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}

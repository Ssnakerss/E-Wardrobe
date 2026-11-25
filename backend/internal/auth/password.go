package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/argon2"
)

const (
	timeParam    = 1
	memory       = 64 * 1024
	threads      = 4
	keyLength    = 32
	saltLength   = 16
)

var errInvalidHash = errors.New("invalid hash format")

// HashPassword hashes a password using Argon2id.
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		timeParam, memory, threads, uint32(keyLength),
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := "$argon2id$v=19$m=65536,t=1,p=4$" + b64Salt + "$" + b64Hash

	return encoded, nil
}

// VerifyPassword checks a password against a hash.
func VerifyPassword(password, hash string) bool {
	salt, hashBytes, err := parseHash(hash)
	if err != nil {
		return false
	}

	keyLength := len(hashBytes)
	computed := argon2.IDKey([]byte(password), salt, timeParam, memory, threads, uint32(keyLength))

	return constantTimeCompare(hashBytes, computed)
}

func parseHash(hash string) ([]byte, []byte, error) {
	parts := []string{}
	remaining := hash
	for len(remaining) > 0 {
		idx := -1
		for i := 0; i < len(remaining); i++ {
			if remaining[i] == '$' {
				idx = i
			}
		}
		if idx == -1 {
			break
		}
		parts = append(parts, remaining[:idx])
		remaining = remaining[idx+1:]
	}

	if len(parts) < 5 {
		return nil, nil, errInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, err
	}

	hashBytes, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, err
	}

	return salt, hashBytes, nil
}

func constantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var v uint8
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}
	return v == 0
}

// GenerateRandomHex generates a random hex string with the given byte count.
func GenerateRandomHex(byteCount int) (string, error) {
	b := make([]byte, byteCount)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashToken returns the SHA-256 hex digest of the given token.
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

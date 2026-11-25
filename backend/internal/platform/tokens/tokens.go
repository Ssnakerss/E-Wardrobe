package tokens

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateRandomToken produces a URL-safe random token with the specified
// number of random bytes. The returned string contains at least that many
// bytes of entropy encoded in base64url.
func GenerateRandomToken(byteCount int) (string, error) {
	b := make([]byte, byteCount)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

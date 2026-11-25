package auth

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/argon2"
)

func TestPasswordDebug(t *testing.T) {
	pw := "test-password"

	// Hash
	hashStr, err := HashPassword(pw)
	require.NoError(t, err)
	t.Logf("Hash string: %s", hashStr)

	// Parse using the fixed logic
	parts := []string{}
	remaining := hashStr
	for len(remaining) > 0 {
		idx := -1
		for i := 0; i < len(remaining); i++ {
			if remaining[i] == '$' {
				idx = i
				break
			}
		}
		if idx == -1 {
			parts = append(parts, remaining)
			break
		}
		parts = append(parts, remaining[:idx])
		remaining = remaining[idx+1:]
	}

	t.Logf("Parts count: %d", len(parts))
	for i, p := range parts {
		t.Logf("  parts[%d] = %q (len=%d)", i, p, len(p))
	}

	require.GreaterOrEqual(t, len(parts), 6, "need at least 6 parts")

	saltB64 := parts[4]
	hashB64 := parts[5]

	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	require.NoError(t, err)
	t.Logf("Salt (hex): %s", hex.EncodeToString(salt))

	hashBytes, err := base64.RawStdEncoding.DecodeString(hashB64)
	require.NoError(t, err)
	t.Logf("HashBytes (hex): %s", hex.EncodeToString(hashBytes))

	// Recompute
	keyLength := len(hashBytes)
	computed := argon2.IDKey([]byte(pw), salt, timeParam, memory, threads, uint32(keyLength))
	t.Logf("Computed (hex): %s", hex.EncodeToString(computed))

	t.Logf("Match: %v", constantTimeCompare(hashBytes, computed))
	require.True(t, constantTimeCompare(hashBytes, computed))
}

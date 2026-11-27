package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests for pure functions: HashPassword, VerifyPassword, GenerateRandomHex, HashToken

func TestHashPassword_DifferentPasswords(t *testing.T) {
	h1, err := HashPassword("password1")
	require.NoError(t, err)
	require.NotEmpty(t, h1)
	require.Contains(t, h1, "$argon2id$")

	h2, err := HashPassword("password2")
	require.NoError(t, err)
	require.NotEmpty(t, h2)
	require.NotEqual(t, h1, h2)
}

func TestHashPassword_SamePasswordDifferentHashes(t *testing.T) {
	h1, err := HashPassword("12qwasZX!@")
	require.NoError(t, err)

	h2, err := HashPassword("12qwasZX!@")
	require.NoError(t, err)

	// Hashes should differ due to random salt
	require.NotEqual(t, h1, h2)
	t.Log(h1)
}

func TestHashPassword_VerifyCorrectPassword(t *testing.T) {
	pw := "mySecurePassword123!"
	hash, err := HashPassword(pw)
	require.NoError(t, err)

	require.True(t, VerifyPassword(pw, hash))
}

func TestHashPassword_VerifyWrongPassword(t *testing.T) {
	hash, err := HashPassword("correct-password")
	require.NoError(t, err)

	require.False(t, VerifyPassword("wrong-password", hash))
}

func TestHashPassword_VerifyEmptyHash(t *testing.T) {
	require.False(t, VerifyPassword("any", ""))
}

func TestHashPassword_VerifyGarbageHash(t *testing.T) {
	require.False(t, VerifyPassword("any", "not-a-valid-hash"))
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	hash, err := HashPassword("")
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	require.True(t, VerifyPassword("", hash))
	require.False(t, VerifyPassword("not-empty", hash))
}

func TestHashPassword_VeryShortPassword(t *testing.T) {
	hash, err := HashPassword("a")
	require.NoError(t, err)

	require.True(t, VerifyPassword("a", hash))
	require.False(t, VerifyPassword("ab", hash))
}

func TestHashPassword_LongPassword(t *testing.T) {
	longPw := ""
	for i := 0; i < 10000; i++ {
		longPw += "a"
	}

	hash, err := HashPassword(longPw)
	require.NoError(t, err)

	require.True(t, VerifyPassword(longPw, hash))
}

func TestGenerateRandomHex_DifferentEachTime(t *testing.T) {
	h1, err := GenerateRandomHex(32)
	require.NoError(t, err)
	require.Len(t, h1, 64) // 32 bytes = 64 hex chars

	h2, err := GenerateRandomHex(32)
	require.NoError(t, err)
	require.NotEqual(t, h1, h2)
}

func TestGenerateRandomHex_ByteCounts(t *testing.T) {
	for _, bc := range []int{1, 4, 16, 32, 64} {
		h, err := GenerateRandomHex(bc)
		require.NoError(t, err)
		require.Len(t, h, bc*2)
	}
}

func TestHashToken_Deterministic(t *testing.T) {
	token := "test-token-12345"
	h1 := HashToken(token)
	h2 := HashToken(token)
	require.Equal(t, h1, h2)
}

func TestHashToken_DifferentTokens(t *testing.T) {
	h1 := HashToken("token-a")
	h2 := HashToken("token-b")
	require.NotEqual(t, h1, h2)
}

func TestHashToken_ReturnsHex(t *testing.T) {
	h := HashToken("test")
	require.Len(t, h, 64) // SHA-256 = 32 bytes = 64 hex chars

	// Verify all chars are valid hex
	for _, c := range h {
		require.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'), "unexpected char %c", c)
	}
}

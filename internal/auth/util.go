package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// sha256hex returns the lowercase hex-encoded SHA-256 digest of s.
func sha256hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// randomHex returns n random bytes from crypto/rand, hex-encoded (2*n chars).
func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

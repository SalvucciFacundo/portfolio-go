package auth

import "crypto/subtle"

// GenerateCSRFToken returns a random 32-byte token as lowercase hex.
func GenerateCSRFToken() (string, error) {
	return randomHex(32)
}

// ValidCSRFToken compares two CSRF tokens in constant time to avoid timing
// attacks on the comparison.
func ValidCSRFToken(provided, stored string) bool {
	return subtle.ConstantTimeCompare([]byte(provided), []byte(stored)) == 1
}

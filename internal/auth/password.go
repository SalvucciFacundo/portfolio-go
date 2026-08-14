package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2idTime    = 3
	argon2idMemory  = 64 * 1024
	argon2idThreads = 4
	argon2idKeyLen  = 32
	argon2idSaltLen = 16
)

// HashPassword derives an argon2id hash of the password and returns it as a
// PHC-formatted string.
func HashPassword(password string) (string, error) {
	salt := make([]byte, argon2idSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("auth: generating salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argon2idTime, argon2idMemory, argon2idThreads, argon2idKeyLen)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argon2idMemory,
		argon2idTime,
		argon2idThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

// VerifyPassword checks a password against an encoded argon2id PHC string.
func VerifyPassword(encoded, password string) (bool, error) {
	params, salt, hash, err := decodePHC(encoded)
	if err != nil {
		return false, err
	}

	candidate := argon2.IDKey([]byte(password), salt, params.time, params.memory, params.threads, uint32(len(hash)))

	return subtle.ConstantTimeCompare(candidate, hash) == 1, nil
}

type argon2Params struct {
	memory  uint32
	time    uint32
	threads uint8
}

func decodePHC(encoded string) (argon2Params, []byte, []byte, error) {
	var params argon2Params

	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return params, nil, nil, errors.New("auth: invalid PHC format: expected $argon2id$v=<ver>$m=<mem>,t=<time>,p=<threads>$<salt>$<hash>")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return params, nil, nil, fmt.Errorf("auth: parsing argon2 version: %w", err)
	}
	if version != argon2.Version {
		return params, nil, nil, fmt.Errorf("auth: unsupported argon2 version %d", version)
	}

	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.memory, &params.time, &params.threads); err != nil {
		return params, nil, nil, fmt.Errorf("auth: parsing argon2 parameters: %w", err)
	}
	if params.memory == 0 || params.time == 0 || params.threads == 0 {
		return params, nil, nil, errors.New("auth: invalid argon2 parameters: m, t and p must be greater than zero")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return params, nil, nil, fmt.Errorf("auth: decoding salt: %w", err)
	}
	if len(salt) < 8 {
		return params, nil, nil, fmt.Errorf("auth: salt too short (%d bytes)", len(salt))
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return params, nil, nil, fmt.Errorf("auth: decoding hash: %w", err)
	}
	if len(hash) < 16 {
		return params, nil, nil, fmt.Errorf("auth: hash too short (%d bytes)", len(hash))
	}

	return params, salt, hash, nil
}

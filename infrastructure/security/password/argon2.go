package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2 returns an arogon2id based password encoding/verifier startegy.
func Argon2() Strategy {
	return &argon2Strategy{
		memory:      64 * 1024,
		iterations:  1,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
	}
}

// -----------------------------------------------------------------------------

type argon2Strategy struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func (s *argon2Strategy) Hash(input string) (string, error) {
	// Generate random bytes as salt
	salt := make([]byte, s.saltLength)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		return "", fmt.Errorf("argon2: unable to generate salt")
	}

	// Use Argon2id
	hash := argon2.IDKey([]byte(input), salt, s.iterations, s.memory, s.parallelism, s.keyLength)

	// Hash is self-encoded.
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		s.memory,
		s.iterations,
		s.parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func (s *argon2Strategy) Verify(plain, secret string) error {
	// Extract argon2 parameters
	hashParts := strings.Split(secret, "$")
	if len(hashParts) != 6 {
		return fmt.Errorf("argon2: invalid secret encoding, missing parts: %w", ErrInvalidSecret)
	}

	// Retrieve salt
	salt, err := base64.RawStdEncoding.DecodeString(hashParts[4])
	if err != nil {
		return fmt.Errorf("argon2: invalid secret encoding, unable to decode salt: %w", ErrInvalidSecret)
	}

	// Retrive hash
	decodedHash, err := base64.RawStdEncoding.DecodeString(hashParts[5])
	if err != nil {
		return fmt.Errorf("argon2: invalid secret encoding, unable to decode hash: %w", ErrInvalidSecret)
	}

	// Compute hash
	hashToCompare := argon2.IDKey([]byte(plain), salt, s.iterations, s.memory, s.parallelism, uint32(len(decodedHash)))

	// Constant-time comparison
	if subtle.ConstantTimeCompare(decodedHash, hashToCompare) == 1 {
		return nil
	}

	// Default to error
	return ErrInvalidSecret
}

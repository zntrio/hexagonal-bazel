package password

import "errors"

type Hasher interface {
	Hash(plain string) (string, error)
}

type Verifier interface {
	Verify(given, secret string) error
}

type Strategy interface {
	Hasher
	Verifier
}

// ErrInvalidSecret is raised when the given and the sotred secret doesn't
// match
var ErrInvalidSecret = errors.New("invalid secret")

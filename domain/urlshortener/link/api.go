package link

import (
	"context"
	"errors"
)

// ID defines the link domain identity type.
type ID string

// Link represents the shortened link contract.
type Link interface {
	// GetID returns the shortened url identifier.
	GetID() ID
	// GetURL returns the original shortened url.
	GetURL() string
	// GetSecretHash returns the secret hash value for verification.
	GetSecretHash() string
}

var (
	// ErrUnableToSaveLink is raised when an error occured during the link
	// domain saving in the persistence.
	ErrUnableToSaveLink = errors.New("unable to save link")

	// ErrLinkNotFound is raised when the persistence query returned no result.
	ErrLinkNotFound = errors.New("link not found")

	// ErrInvalidSecret is raised when the given secret doesn't match
	// the stored one.
	ErrInvalidSecret = errors.New("invalid secret")
)

// -----------------------------------------------------------------------------

// Resolver represents Link domain object resolver from repository.
type Resolver interface {
	// GetByID returns the Link domain object matching the given identifier.
	GetByID(ctx context.Context, id ID) (Link, error)
}

// Writer represent Link domain object write operations in the repository.
type Writer interface {
	// Save a shortened url domain object in the repository.
	Save(ctx context.Context, m Link) error
}

type Repository interface {
	Resolver
	Writer
}

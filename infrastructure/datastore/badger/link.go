package badger

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/timshannon/badgerhold/v4"
	"golang.org/x/crypto/blake2b"

	"zntr.io/hexagonal-bazel/domain/urlshortener/link"
	"zntr.io/hexagonal-bazel/pkg/types"
)

var (
	seedKey = []byte("I<%|J/Fs\\(%rhU4kE!v+W<:|%-`@aZw7uyb*3f8I?pK'9=w&,lPL6Ds#bI?PW.$")
)

// Links returns a badger implementation of the repository.
func Links(db *badgerhold.Store) link.Repository {
	return &linkRepository{
		db: db,
	}
}

// -----------------------------------------------------------------------------

type linkRepository struct {
	db *badgerhold.Store
}

// -----------------------------------------------------------------------------

type linkEntity struct {
	ID         string     `json:"id"`
	URL        string     `json:"url"`
	SecretHash string     `json:"secret_hash"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
}

var _ link.Link = (*linkEntity)(nil)

func (e *linkEntity) GetID() link.ID           { return link.ID(e.ID) }
func (e *linkEntity) GetURL() string           { return e.URL }
func (e *linkEntity) GetSecretHash() string    { return e.SecretHash }
func (e *linkEntity) GetCreatedAt() time.Time  { return e.CreatedAt }
func (d *linkEntity) GetExpiresAt() *time.Time { return d.ExpiresAt }
func (d *linkEntity) IsProtected() bool        { return d.SecretHash != "" }
func (d *linkEntity) IsExpired(ref time.Time) bool {
	return d.ExpiresAt != nil && ref.After(*d.ExpiresAt)
}

// -----------------------------------------------------------------------------

func (r *linkRepository) GetByID(ctx context.Context, id link.ID) (link.Link, error) {
	// Create a transaction
	tx := r.db.Badger().NewTransaction(false)

	// Compute ID due to the fact that this information could be given by the
	// user. We have to enforce the confidentiality of the ID in the case the
	// user registered it with sensitive data.
	nid := r.computeID(id)

	var result linkEntity
	if err := r.db.TxGet(tx, nid, &result); err != nil {
		if errors.Is(err, badgerhold.ErrNotFound) {
			return nil, link.ErrLinkNotFound
		}
		return nil, fmt.Errorf("badger: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("badger: transaction error - %v: %w", err, link.ErrUnableToSaveLink)
	}

	// No error
	return &result, nil
}

func (r *linkRepository) Save(ctx context.Context, domain link.Link) error {
	// Check arguments
	if types.IsNil(domain) {
		return errors.New("badger: unable to save a nil link")
	}

	// Compute ID due to the fact that this information could be given by the
	// user. We have to enforce the confidentiality of the ID in the case the
	// user registered it with sensitive data.
	id := r.computeID(domain.GetID())

	// Convert to entity
	entity := &linkEntity{
		ID:         id,
		URL:        domain.GetURL(),
		SecretHash: domain.GetSecretHash(),
		CreatedAt:  domain.GetCreatedAt(),
		ExpiresAt:  domain.GetExpiresAt(),
	}

	// Create a transaction
	tx := r.db.Badger().NewTransaction(true)

	// Insert or update.
	if err := r.db.TxUpsert(tx, id, entity); err != nil {
		return fmt.Errorf("badger: %v: %w", err, link.ErrUnableToSaveLink)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("badger: transaction error - %v: %w", err, link.ErrUnableToSaveLink)
	}

	// No error
	return nil
}

// -----------------------------------------------------------------------------

func (r *linkRepository) computeID(id link.ID) string {
	// Prepare keyed blake2b
	h, err := blake2b.New256(seedKey)
	if err != nil {
		panic(err)
	}

	// Add domain separation to ensure unique result on key reuse.
	// Domain || 0x00 || ID
	h.Write([]byte("urlshortener:v1:link"))
	h.Write([]byte{0x00})
	h.Write([]byte(id))

	// Encode result as raw base64 url.
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

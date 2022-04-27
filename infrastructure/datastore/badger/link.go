package badger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"zntr.io/hexagonal-bazel/domain/urlshortener/link"
	"zntr.io/hexagonal-bazel/pkg/types"

	"github.com/timshannon/badgerhold/v4"
)

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
	ID         string    `json:"id"`
	URL        string    `json:"url"`
	SecretHash string    `json:"secret_hash"`
	CreatedAt  time.Time `json:"created_at"`
}

var _ link.Link = (*linkEntity)(nil)

func (e *linkEntity) GetID() link.ID          { return link.ID(e.ID) }
func (e *linkEntity) GetURL() string          { return e.URL }
func (e *linkEntity) GetSecretHash() string   { return e.SecretHash }
func (e *linkEntity) GetCreatedAt() time.Time { return e.CreatedAt }
func (d *linkEntity) IsProtected() bool       { return d.SecretHash != "" }

// -----------------------------------------------------------------------------

func (r *linkRepository) GetByID(ctx context.Context, id link.ID) (link.Link, error) {
	// Create a transaction
	tx := r.db.Badger().NewTransaction(false)

	var result linkEntity
	if err := r.db.TxGet(tx, id, &result); err != nil {
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

	// Convert to entity
	entity := &linkEntity{
		ID:         string(domain.GetID()),
		URL:        domain.GetURL(),
		SecretHash: domain.GetSecretHash(),
		CreatedAt:  domain.GetCreatedAt(),
	}

	// Create a transaction
	tx := r.db.Badger().NewTransaction(true)

	// Insert or update.
	if err := r.db.TxUpsert(tx, entity.GetID(), entity); err != nil {
		return fmt.Errorf("badger: %v: %w", err, link.ErrUnableToSaveLink)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("badger: transaction error - %v: %w", err, link.ErrUnableToSaveLink)
	}

	// No error
	return nil
}

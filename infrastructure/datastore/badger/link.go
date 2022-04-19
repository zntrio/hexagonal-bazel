package badger

import (
	"context"
	"errors"
	"fmt"

	"zntr.io/hexagonal-bazel/domain/urlshortener/link"

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
	ID  string `json:"id"`
	URL string `json:"url"`
}

var _ link.Link = (*linkEntity)(nil)

func (e *linkEntity) GetID() link.ID { return link.ID(e.ID) }
func (e *linkEntity) GetURL() string { return e.URL }

// -----------------------------------------------------------------------------

func (r *linkRepository) GetByID(ctx context.Context, id link.ID) (link.Link, error) {
	var result linkEntity
	if err := r.db.Get(id, &result); err != nil {
		if errors.Is(err, badgerhold.ErrNotFound) {
			return nil, link.ErrLinkNotFound
		}
		return nil, fmt.Errorf("badger: %w", err)
	}

	// No error
	return &result, nil
}

func (r *linkRepository) Save(ctx context.Context, domain link.Link) error {
	// Convert to entity
	entity := &linkEntity{
		ID:  string(domain.GetID()),
		URL: domain.GetURL(),
	}

	// Insert or update.
	if err := r.db.Upsert(entity.GetID(), entity); err != nil {
		return fmt.Errorf("badger: %v: %w", err, link.ErrUnableToSaveLink)
	}

	// No error
	return nil
}

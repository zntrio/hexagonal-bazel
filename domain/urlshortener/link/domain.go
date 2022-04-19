package link

import (
	"sync"
	"time"

	"github.com/teris-io/shortid"
)

var (
	sid  *shortid.Shortid
	once sync.Once
)

func init() {
	once.Do(func() {
		var err error
		sid, err = shortid.New(1, shortid.DefaultABC, uint64(time.Now().Unix()))
		if err != nil {
			panic(err)
		}

		// Assign as default instance
		shortid.SetDefault(sid)
	})
}

// NewID generates a new shortened link public identifier.
func NewID() ID {
	return ID(sid.MustGenerate())
}

// defaultLink implements the Link domain object interface.
type defaultLink struct {
	id  ID
	url string
}

// Compile-time assertion to enforce type implementation.
var _ Link = (*defaultLink)(nil)

func (d *defaultLink) GetID() ID      { return d.id }
func (d *defaultLink) GetURL() string { return d.url }

// -----------------------------------------------------------------------------

// DomainOption declares the functional option type.
type DomainOption func(*defaultLink)

// WithID sets the domain object identifier.
func WithID(id ID) DomainOption {
	return func(dopts *defaultLink) {
		dopts.id = id
	}
}

// WithURL sets the domain object URL property.
func WithURL(url string) DomainOption {
	return func(dopts *defaultLink) {
		dopts.url = url
	}
}

// -----------------------------------------------------------------------------

// New link object domain instance.
func New(opts ...DomainOption) Link {
	def := &defaultLink{}

	// Apply all options
	for _, o := range opts {
		o(def)
	}

	return def
}

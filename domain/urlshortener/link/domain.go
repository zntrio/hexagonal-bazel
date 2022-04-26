package link

// defaultLink implements the Link domain object interface.
type defaultLink struct {
	id         ID
	url        string
	secretHash string
}

// Compile-time assertion to enforce type implementation.
var _ Link = (*defaultLink)(nil)

func (d *defaultLink) GetID() ID             { return d.id }
func (d *defaultLink) GetURL() string        { return d.url }
func (d *defaultLink) GetSecretHash() string { return d.secretHash }
func (d *defaultLink) IsProtected() bool     { return d.secretHash != "" }

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

// WithSecretHash sets the domain object secret hash property.
func WithSecretHash(value string) DomainOption {
	return func(dopts *defaultLink) {
		dopts.secretHash = value
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

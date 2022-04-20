package shortid

import (
	"github.com/teris-io/shortid"

	"zntr.io/hexagonal-bazel/infrastructure/generator"
)

func New(worker uint8) generator.Generator[string] {
	return &shortidGenerator{
		sid: shortid.MustNew(worker, shortid.DefaultABC, 0x1badc0ffee),
	}
}

// -----------------------------------------------------------------------------

type shortidGenerator struct {
	sid *shortid.Shortid
}

func (g *shortidGenerator) Generate() (string, error) {
	return g.sid.Generate()
}

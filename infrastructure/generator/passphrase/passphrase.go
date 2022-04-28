package passphrase

import (
	"fmt"
	"strings"

	"github.com/sethvargo/go-diceware/diceware"

	"zntr.io/hexagonal-bazel/infrastructure/generator"
)

func Diceware(wordCount int) generator.Generator[string] {
	return &dicewareGenerator{
		wordCount: wordCount,
	}
}

// -----------------------------------------------------------------------------

type dicewareGenerator struct {
	wordCount int
}

func (g *dicewareGenerator) Generate() (string, error) {
	// Generate requested number of words using the diceware algorithm.
	list, err := diceware.Generate(g.wordCount)
	if err != nil {
		return "", fmt.Errorf("diceware: unable to generate a passphrase: %w", err)
	}

	// No error
	return strings.Join(list, "-"), nil
}

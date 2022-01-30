package pkg

import (
	"fmt"
	"math/rand"
)

type keyGenerator struct {
	keyPrefix string
	keySpaceLen int
}

func newKeyGenerator(keyPrefix string, keySpaceLen int) *keyGenerator {
	return &keyGenerator{
		keyPrefix: keyPrefix,
		keySpaceLen: keySpaceLen,
	}
}

func (g *keyGenerator) genKey() string {
	return fmt.Sprintf("%s:%d", g.keyPrefix, rand.Intn(g.keySpaceLen))
}

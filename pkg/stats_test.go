package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHistogram(t *testing.T) {
	assert := assert.New(t)

	h := NewHistogram(1, 8)
	h.counters = []uint64{
		2, 4, 9,
		0, 4, 5,
		6, 6,
	}

	h.Add(16)
	assert.Equal(uint64(3), h.unit)

	assert.Equal(uint64(15), h.counters[0])
	assert.Equal(uint64(9), h.counters[1])
	assert.Equal(uint64(12), h.counters[2])

	assert.Equal(uint64(0), h.counters[3])
	assert.Equal(uint64(0), h.counters[4])
	assert.Equal(uint64(1), h.counters[5])

	assert.Equal(uint64(0), h.counters[6])
	assert.Equal(uint64(0), h.counters[7])
}

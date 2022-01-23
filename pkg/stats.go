package pkg

// Histogram implements a high performance way
// to build up a auto-scale histogram.
// This is not thread-safe.
type Histogram struct {
	// For scaling
	unit uint64
	counters []uint64
}

// NewHistogram creates Histogram
func NewHistogram(initUnit, unitNum uint64) *Histogram {
	counters := make([]uint64, 0, unitNum)
	for i := uint64(0); i != unitNum; i++ {
		counters = append(counters, 0)
	}
	return &Histogram{
		unit: initUnit,
		counters: counters,
	}
}

// Add adds a sample
func (h *Histogram) Add(num uint64) {
	currMax := h.unit * uint64(len(h.counters))
	if num >= currMax {
		h.scale(num)
	}

	index := num / h.unit
	h.counters[index] += 1
}

func (h *Histogram) scale(num uint64) {
	currMax := h.unit * uint64(len(h.counters))
	// currMax * scaleFactor - 1 >= num
	scaleFactor := (num + 1 + currMax - 1) / currMax

	countersLen := uint64(len(h.counters))
	for i := uint64(0); i * scaleFactor < countersLen; i++ {
		start := i * scaleFactor
		var s uint64 = 0
		for j := uint64(0); j != scaleFactor && start + j != countersLen; j++ {
			s += h.counters[start + j]
			h.counters[start + j] = 0
		}
		h.counters[i] = s
	}

	h.unit *= scaleFactor
}

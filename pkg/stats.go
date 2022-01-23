package pkg

import "context"


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

func (h *Histogram) GetCount() uint64 {
	var count uint64 = 0
	for _, c := range h.counters {
		count += c
	}
	return count
}

type HistogramResult struct {
	start uint64
	end uint64
	count uint64
}

func (h *Histogram) GetResults() []HistogramResult {
	results := make([]HistogramResult, 0, len(h.counters))
	for i, count := range h.counters {
		i := uint64(i)
		results = append(results, HistogramResult{
			start: i * h.unit,
			end: (i + 1) * h.unit - 1,
			count: count,
		})
	}
	return results
}

func (h *Histogram) addMulti(nums []uint64) {
	for _, num := range nums {
		h.Add(num)
	}
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

type sampleCollector struct {
	histogram *Histogram
	dataC chan []uint64
	stopC chan bool
}

func newSampleCollector(initUnit uint64) *sampleCollector {
	return &sampleCollector{
		histogram: NewHistogram(initUnit, 10),
		dataC: make(chan []uint64, 1024),
	}
}

func (c *sampleCollector) run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-c.stopC:
			close(c.dataC)
			for samples := range c.dataC {
				c.histogram.addMulti(samples)
			}
			return nil
		case samples := <-c.dataC:
			c.histogram.addMulti(samples)
		}
	}
}

func (c *sampleCollector) stop() {
	c.stopC <- true
}

func (c *sampleCollector) add(samples []uint64) {
	c.dataC <- samples
}

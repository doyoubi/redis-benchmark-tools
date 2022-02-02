package pkg

import (
	"fmt"
	"strings"
)

type resultPrinter struct {}

func (d *resultPrinter) output(result *benchmarkResult) error {
	histogram := result.Histogram

	segments := histogram.GetResults()

	var maxCount uint64
	for _, segment := range segments {
		if segment.count > maxCount {
			maxCount = segment.count
		}
	}

	const maxLen uint64 = 30

	fmt.Printf("range\tcount\t\n")
	for _, segment := range segments {
		l := maxLen * segment.count / maxCount
		bar := strings.Repeat("#", int(l))
		fmt.Printf("%d - %d\t%d\t%s\n", segment.start, segment.end, segment.count, bar)
	}

	commands := histogram.GetCount()
	fmt.Printf("commands: %d\n", commands)
	fmt.Printf("duration: %d\n", result.Duration)
	qps := commands * 1000 / uint64(result.Duration.Milliseconds())
	fmt.Printf("qps: %d\n", qps)

	return nil
}

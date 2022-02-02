package pkg

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
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

	const maxLen uint64 = 60

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"start (micro-sec)", "end (micro-sec)", "count", ""})
	for _, segment := range segments {
		l := maxLen * segment.count / maxCount
		bar := strings.Repeat("#", int(l))
		t.AppendRow([]interface{}{segment.start/1000, segment.end/1000, segment.count, bar})
	}
	t.Render()

	commands := histogram.GetCount()
	fmt.Printf("commands: %d\n", commands)
	fmt.Printf("duration: %f ms\n", result.Duration.Seconds())
	qps := commands * 1000 / uint64(result.Duration.Milliseconds())
	fmt.Printf("qps: %d\n", qps)

	return nil
}

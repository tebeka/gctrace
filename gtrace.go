/*
Package gctrace parses Go's GC trace output lines.

See [runtime] & [blog] for information about the format.

[runtime]: https://pkg.go.dev/runtime#hdr-Environment_Variables
[blog]: https://www.ardanlabs.com/blog/2019/05/garbage-collection-in-go-part2-gctraces.html
*/
package gctrace

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// Wall is wall clock statistics.
type Wall struct {
	SweepTermination float64 // Sweep termination stop-the-world (STW)
	MarkAndScan      float64 // Concurrent mark and scan
	MarkTermination  float64 // Mark termination stop-the-world (STW)
}

// CPU is wall CPU times statistics.
type CPU struct {
	SweepTermination float64 // Sweep termination stop-the-world (STW)
	MarkAssist       float64 // GC performed in line with allocation
	MarkBackground   float64 // Background GC time
	MarkIDLE         float64 // Idle GC time
	MarkTermination  float64 // Mark termination stop-the-world (STW)
}

// Trace represent a GC trace line.
type Trace struct {
	Num        int           // The GC number
	Start      time.Duration // Time since program star
	Percentage float64       // Percentage of time spent in GC since program start
	Wall       Wall          // Wall clock times
	CPU        CPU           // CPU times
	Forced     bool          // Forced by a runtime.GC() call
}

var (
	re = regexp.MustCompile(`gc (?P<num>\d+) @(?P<start>\d+\.(\d+)?s)`)
)

func idxOf(name string) int {
	return re.SubexpIndex(name)
}

// Parse parsers a gctrace line.
func Parse(line string) (Trace, error) {
	match := re.FindStringSubmatch(line)
	if match == nil {
		return Trace{}, fmt.Errorf("no match")
	}

	var tr Trace
	var err error
	tr.Num, err = strconv.Atoi(match[idxOf("num")])
	if err != nil {
		return Trace{}, err
	}
	tr.Start, err = time.ParseDuration(match[idxOf("start")])
	if err != nil {
		return Trace{}, err
	}

	return tr, nil
}

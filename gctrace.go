package gctrace

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Wall is wall clock statistics.
type Wall struct {
	SweepTermination time.Duration // Sweep termination stop-the-world (STW)
	MarkAndScan      time.Duration // Concurrent mark and scan
	MarkTermination  time.Duration // Mark termination stop-the-world (STW)
}

// CPU is wall CPU times statistics.
type CPU struct {
	SweepTermination time.Duration // Sweep termination stop-the-world (STW)
	MarkAssist       time.Duration // GC performed in line with allocation
	MarkBackground   time.Duration // Background GC time
	MarkIdle         time.Duration // Idle GC time
	MarkTermination  time.Duration // Mark termination stop-the-world (STW)
}

type Heap struct {
	Before int
	After  int
	Live   int
	Goal   int
}

// Trace represent a GC trace line.
type Trace struct {
	Num        int           // The GC number
	Start      time.Duration // Time since program star
	Percentage float64       // Percentage of time spent in GC since program start
	Wall       Wall          // Wall clock times
	CPU        CPU           // CPU times
	Heap       Heap          // Heap stats
	Cores      int           // Number of processors used
	Forced     bool          // Forced by a runtime.GC() call
}

var (
	// Example:
	// gc 18 @1.824s 13%: 0.030+44+0.015 ms clock, 0.12+29/43/0+0.060 ms cpu, 173->203->101 MB, 203 MB goal, 0 MB stacks, 0 MB globals, 4 P
	re = regexp.MustCompile(`gc (?P<num>\d+) @(?P<start>\d+\.(\d+)?s) (?P<prec>[^%]+)%: (?P<clock>[^ ]+) ms clock, (?P<cpu>[^ ]+) ms cpu, (?P<before>\d+)->(?P<after>\d+)->(?P<live>\d+) MB, (?P<goal>\d+) MB goal,.*?(?P<cores>\d+) P.*?(?P<forced>\(forced\))?`)
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

	tr.Percentage, err = strconv.ParseFloat(match[idxOf("prec")], 64)
	if err != nil {
		return Trace{}, err
	}

	if err := parseWall(match[idxOf("clock")], &tr.Wall); err != nil {
		return Trace{}, err
	}

	if err := parseCPU(match[idxOf("cpu")], &tr.CPU); err != nil {
		return Trace{}, err
	}

	// TODO: Multiply by MB?
	tr.Heap.Before, err = strconv.Atoi(match[idxOf("before")])
	if err != nil {
		return Trace{}, err
	}

	tr.Heap.After, err = strconv.Atoi(match[idxOf("after")])
	if err != nil {
		return Trace{}, err
	}

	tr.Heap.Live, err = strconv.Atoi(match[idxOf("live")])
	if err != nil {
		return Trace{}, err
	}

	tr.Heap.Goal, err = strconv.Atoi(match[idxOf("goal")])
	if err != nil {
		return Trace{}, err
	}

	tr.Cores, err = strconv.Atoi(match[idxOf("cores")])
	if err != nil {
		return Trace{}, err
	}

	if len(match[idxOf("forced")]) > 0 {
		tr.Forced = true
	}

	return tr, nil
}

func parseWall(s string, w *Wall) error {
	// 0.033+44+0.015
	fields := strings.Split(s, "+")
	if len(fields) != 3 {
		return fmt.Errorf("bad clock: %q", s)
	}

	var err error
	w.SweepTermination, err = time.ParseDuration(fields[0] + "ms")
	if err != nil {
		return fmt.Errorf("bad clock: %q - %s", s, err)
	}

	w.MarkAndScan, err = time.ParseDuration(fields[1] + "ms")
	if err != nil {
		return fmt.Errorf("bad clock: %q - %s", s, err)
	}

	w.MarkTermination, err = time.ParseDuration(fields[2] + "ms")
	if err != nil {
		return fmt.Errorf("bad clock: %q - %s", s, err)
	}

	return nil
}

func parseCPU(s string, cpu *CPU) error {
	// 0.12+29/43/0+0.060
	fields := strings.FieldsFunc(s, cpuSplit)
	if len(fields) != 5 {
		return fmt.Errorf("bad cpu: %q", s)
	}

	var err error
	cpu.SweepTermination, err = time.ParseDuration(fields[0] + "ms")
	if err != nil {
		return fmt.Errorf("%q: %s", s, err)
	}

	cpu.MarkAssist, err = time.ParseDuration(fields[1] + "ms")
	if err != nil {
		return fmt.Errorf("%q: %s", s, err)
	}

	cpu.MarkBackground, err = time.ParseDuration(fields[2] + "ms")
	if err != nil {
		return fmt.Errorf("%q: %s", s, err)
	}

	cpu.MarkIdle, err = time.ParseDuration(fields[3] + "ms")
	if err != nil {
		return fmt.Errorf("%q: %s", s, err)
	}

	cpu.MarkTermination, err = time.ParseDuration(fields[4] + "ms")
	if err != nil {
		return fmt.Errorf("%q: %s", s, err)
	}

	return nil
}

func cpuSplit(r rune) bool {
	return r == '/' || r == '+'
}

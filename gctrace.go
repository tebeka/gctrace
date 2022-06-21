package gctrace

import (
	"bufio"
	"fmt"
	"io"
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
	// TODO
	// Forced     bool          // Forced by a runtime.GC() call
}

var (
	// Example:
	// gc 18 @1.824s 13%: 0.030+44+0.015 ms clock, 0.12+29/43/0+0.060 ms cpu, 173->203->101 MB, 203 MB goal, 0 MB stacks, 0 MB globals, 4 P
	re = regexp.MustCompile(`gc (?P<num>\d+) @(?P<start>\d+\.(\d+)?s) (?P<prec>[^%]+)%: (?P<clock>[^ ]+) ms clock, (?P<cpu>[^ ]+) ms cpu, (?P<before>\d+)->(?P<after>\d+)->(?P<live>\d+) MB, (?P<goal>\d+) MB goal,.*?(?P<cores>\d+) P`)
)

func sub(name string, match []string) string {
	i := re.SubexpIndex(name)
	if i == -1 {
		panic(fmt.Sprintf("unknown sub-expression: %q", name))
	}
	return match[i]
}

// Unmarshal parsers a gctrace line.
func Unmarshal(line string, tr *Trace) error {
	match := re.FindStringSubmatch(line)
	if match == nil {
		return fmt.Errorf("no match")
	}

	var err error
	tr.Num, err = strconv.Atoi(sub("num", match))
	if err != nil {
		return err
	}

	tr.Start, err = time.ParseDuration(sub("start", match))
	if err != nil {
		return err
	}

	tr.Percentage, err = strconv.ParseFloat(sub("prec", match), 64)
	if err != nil {
		return err
	}

	if err := parseWall(sub("clock", match), &tr.Wall); err != nil {
		return err
	}

	if err := parseCPU(sub("cpu", match), &tr.CPU); err != nil {
		return err
	}

	// TODO: Multiply by MB?
	tr.Heap.Before, err = strconv.Atoi(sub("before", match))
	if err != nil {
		return err
	}

	tr.Heap.After, err = strconv.Atoi(sub("after", match))
	if err != nil {
		return err
	}

	tr.Heap.Live, err = strconv.Atoi(sub("live", match))
	if err != nil {
		return err
	}

	tr.Heap.Goal, err = strconv.Atoi(sub("goal", match))
	if err != nil {
		return err
	}

	tr.Cores, err = strconv.Atoi(sub("cores", match))
	if err != nil {
		return err
	}

	return nil
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

// Scanner is a trace scanner
type Scanner struct {
	lnum int
	s    *bufio.Scanner
	err  error
	tr   Trace
}

// NewScanner returns a new Scanner over r.
func NewScanner(r io.Reader) *Scanner {
	sc := Scanner{
		s: bufio.NewScanner(r),
	}
	return &sc
}

// Scan advances to the next trace.
func (s *Scanner) Scan() bool {
	if s.err != nil {
		return false
	}

	if !s.s.Scan() {
		return false
	}

	s.err = s.s.Err()
	if s.err != nil {
		return false
	}
	s.lnum++

	// Skip:
	// # command-line-arguments
	if strings.HasPrefix(s.s.Text(), "#") {
		return s.Scan()
	}

	s.tr = Trace{}
	s.err = Unmarshal(s.s.Text(), &s.tr)
	if s.err != nil {
		s.tr = Trace{}
		return false
	}
	return true
}

// Err returns the last error.
func (s *Scanner) Err() error {
	return s.err
}

// Line returns the current line.
func (s *Scanner) Line() string {
	return s.s.Text()
}

// LineNum returns the current line number.
func (s *Scanner) LineNum() int {
	return s.lnum
}

// Trace returns the current parsed trace
func (s *Scanner) Trace() Trace {
	return s.tr
}

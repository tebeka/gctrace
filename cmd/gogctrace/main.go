package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/tebeka/gctrace"
)

const usage = `usage: gogctrace [FILE]
Convert Go's GC trace output lines to JSON (one per line).

`

func main() {
	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	fileName := "<stdin>"

	var r io.Reader
	switch flag.NArg() {
	case 0:
		r = os.Stdin
	case 1:
		fileName = flag.Arg(0)
		file, err := os.Open(fileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: can't open %q - %s", fileName, err)
			os.Exit(1)
		}
		defer file.Close()
		r = file
	default:
		fmt.Fprintln(os.Stderr, "error: can't wrong number of arguments")
		os.Exit(1)
	}

	s := gctrace.NewScanner(r)
	for s.Scan() {
		data, err := traceJSON(s.Trace())
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s:%d - %s", fileName, s.LineNum(), err)
			os.Exit(1)
		}
		os.Stdout.Write(data)
		os.Stdout.Write([]byte{'\n'})
	}
	if err := s.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s:%d - %s", fileName, s.LineNum(), err)
		os.Exit(1)
	}
}

func traceJSON(t gctrace.Trace) ([]byte, error) {
	obj := map[string]any{
		"num":        t.Num,
		"start":      t.Start,
		"percentage": t.Percentage,
		"wall": map[string]any{
			"sweep_termination": t.Wall.SweepTermination,
			"mark_and_scan":     t.Wall.MarkAndScan,
			"mark_termination":  t.Wall.MarkTermination,
		},
		"cpu": map[string]any{
			"sweep_termination": t.CPU.SweepTermination,
			"mark_assist":       t.CPU.MarkAssist,
			"mark_background":   t.CPU.MarkBackground,
			"mark_idle":         t.CPU.MarkIdle,
			"mark_termination":  t.CPU.MarkTermination,
		},
		"heap": map[string]any{
			"before": t.Heap.Before,
			"after":  t.Heap.After,
			"live":   t.Heap.Live,
			"goal":   t.Heap.Goal,
		},
		"cores": t.Cores,
	}
	return json.Marshal(obj)
}

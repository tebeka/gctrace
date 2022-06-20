package gctrace

import (
	"os"
	"testing"
	"time"
)

func TestTrace(t *testing.T) {
	line := `gc 18 @1.824s 13%: 0.030+44+0.015 ms clock, 0.12+29/43/0+0.060 ms cpu, 173->203->101 MB, 203 MB goal, 0 MB stacks, 0 MB globals, 4 P`

	var tr Trace
	if err := Unmarshal(line, &tr); err != nil {
		t.Fatal(err)
	}

	expected := Trace{
		Num:        18,
		Start:      time.Second + 824*time.Millisecond,
		Percentage: 13,
		Wall: Wall{
			SweepTermination: 30 * time.Microsecond,
			MarkAndScan:      44 * time.Millisecond,
			MarkTermination:  15 * time.Microsecond,
		},
		CPU: CPU{
			SweepTermination: 120 * time.Microsecond,
			MarkAssist:       29 * time.Millisecond,
			MarkBackground:   43 * time.Millisecond,
			MarkIdle:         0,
			MarkTermination:  60 * time.Microsecond,
		},
		Heap: Heap{
			Before: 173,
			After:  203,
			Live:   101,
			Goal:   203,
		},
		Cores: 4,
	}

	if expected != tr {
		t.Logf("tr: %+v\n", tr)
		t.Logf("expected: %+v\n", expected)
		t.Fatal()
	}
}

func TestScanner(t *testing.T) {
	file, err := os.Open("testdata/map.gctrace")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	s := NewScanner(file)
	count := 0
	for s.Next() {
		count++
	}
	if err := s.Err(); err != nil {
		t.Fatal(err)
	}

	const nr = 95
	if count != nr {
		t.Fatalf("count: expected %d, got %d", nr, count)
	}

	const nl = 97
	if lnum := s.Line(); lnum != nl {
		t.Fatalf("lines: expected %d, got %d", nl, lnum)
	}
}

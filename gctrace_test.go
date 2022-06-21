package gctrace

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTrace(t *testing.T) {
	require := require.New(t)
	line := `gc 18 @1.824s 13%: 0.030+44+0.015 ms clock, 0.12+29/43/0+0.060 ms cpu, 173->203->101 MB, 203 MB goal, 0 MB stacks, 0 MB globals, 4 P`

	var tr Trace
	err := Unmarshal(line, &tr)
	require.NoError(err)

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

	require.Equal(expected, tr)
}

func TestScanner(t *testing.T) {
	require := require.New(t)
	file, err := os.Open("testdata/map.gctrace")
	require.NoError(err)
	defer file.Close()

	s := NewScanner(file)
	count := 0
	for s.Next() {
		count++
	}
	require.NoError(s.Err())
	require.Equal(95, count, "trace count")
	require.Equal(97, s.LineNum(), "line count")
}

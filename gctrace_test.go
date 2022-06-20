package gctrace

import (
	"testing"
	"time"
)

func TestTrace(t *testing.T) {
	line := `gc 18 @1.824s 13%: 0.030+44+0.015 ms clock, 0.12+29/43/0+0.060 ms cpu, 173->203->101 MB, 203 MB goal, 0 MB stacks, 0 MB globals, 4 P`

	tr, err := Parse(line)
	if err != nil {
		t.Fatal(err)
	}

	expected := Trace{
		Num:   18,
		Start: time.Second + 824*time.Millisecond,
	}

	if expected != tr {
		t.Fatal(tr)
	}
}

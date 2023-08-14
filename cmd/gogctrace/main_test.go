package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHelp(t *testing.T) {
	exe := build(t)
	var buf bytes.Buffer
	cmd := exec.Command(exe, "-h")
	cmd.Stderr = &buf
	require.NoError(t, cmd.Run(), "run")
	require.Contains(t, buf.String(), "usage:", "usage")
}

func TestExe(t *testing.T) {
	exe := build(t)
	fileName := "../../testdata/map.gctrace"
	numTraces := traceCount(t, fileName)

	file, err := os.Open(fileName)
	require.NoErrorf(t, err, "open %q", fileName)
	defer file.Close()

	cmd := exec.Command(exe)
	cmd.Stdin = file
	out, err := cmd.Output()
	require.NoError(t, err, "run")

	s := bufio.NewScanner(bytes.NewReader(out))
	lnum := 0
	for s.Scan() {
		lnum++
		var m map[string]any
		err := json.Unmarshal(s.Bytes(), &m)
		require.NoError(t, err, "unmarshal JSON at line %d - %s", lnum, err)
	}
	require.NoError(t, s.Err(), "scan")

	require.Equal(t, numTraces, lnum, "number of lines")
}

func build(t *testing.T) string {
	exe := fmt.Sprintf("%s/gogctrace", t.TempDir())
	err := exec.Command("go", "build", "-o", exe).Run()
	require.NoError(t, err, "build")
	return exe
}

func traceCount(t *testing.T, fileName string) int {
	file, err := os.Open(fileName)
	require.NoError(t, err, "open")
	defer file.Close()

	count := 0
	s := bufio.NewScanner(file)
	for s.Scan() {
		if s.Text()[0] != '#' {
			count++
		}
	}
	require.NoError(t, s.Err(), "scan")
	return count
}

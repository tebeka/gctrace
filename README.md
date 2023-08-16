# gctrace

**NOT READY YET, USE AT YOUR OWN RISK!**

Parse `GODEBUG=gctrace=1` output lines. 
There's also a utility to convert the output to JSON, see [here][pkg] and download from [here][rel].

[pkg]: https://pkg.go.dev/github.com/tebeka/gctrace/cmd/gogctrace
[rel]: https://github.com/tebeka/gctrace/releases

Examples (see [example_test.go](example_test.go)):

```go
func ExampleUnmarshal() {
	line := `gc 18 @1.824s 13%: 0.030+44+0.015 ms clock, 0.12+29/43/0+0.060 ms cpu, 173->203->101 MB, 203 MB goal, 0 MB stacks, 0 MB globals, 4 P`
	var tr gctrace.Trace
	if err := gctrace.Unmarshal(line, &tr); err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	fmt.Printf("%+v\n", tr)

	// Output:
	// {Num:18 Start:1.824s Percentage:13 Wall:{SweepTermination:30µs MarkAndScan:44ms MarkTermination:15µs} CPU:{SweepTermination:120µs MarkAssist:29ms MarkBackground:43ms MarkIdle:0s MarkTermination:60µs} Heap:{Before:173 After:203 Live:101 Goal:203} Cores:4}
}

func ExampleScan() {
	lines := `gc 1 @0.009s 4%: 0.072+1.5+0.004 ms clock, 0.28+0.71/1.1/0.050+0.016 ms cpu, 4->4->0 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 4 P
gc 2 @0.014s 6%: 0.066+0.80+0.003 ms clock, 0.26+1.0/0.62/0+0.012 ms cpu, 4->4->0 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 4 P
gc 3 @0.018s 7%: 0.10+0.60+0.013 ms clock, 0.42+1.0/0.55/0+0.055 ms cpu, 4->4->0 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 4 P`
	r := strings.NewReader(lines)

	s := gctrace.NewScanner(r)
	for s.Scan() {
		fmt.Printf("%d: %+v\n", s.LineNum(), s.Trace())
	}
	if err := s.Err(); err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	// Output:
	// 1: {Num:1 Start:9ms Percentage:4 Wall:{SweepTermination:72µs MarkAndScan:1.5ms MarkTermination:4µs} CPU:{SweepTermination:280µs MarkAssist:710µs MarkBackground:1.1ms MarkIdle:50µs MarkTermination:16µs} Heap:{Before:4 After:4 Live:0 Goal:4} Cores:4}
	// 2: {Num:2 Start:14ms Percentage:6 Wall:{SweepTermination:66µs MarkAndScan:800µs MarkTermination:3µs} CPU:{SweepTermination:260µs MarkAssist:1ms MarkBackground:620µs MarkIdle:0s MarkTermination:12µs} Heap:{Before:4 After:4 Live:0 Goal:4} Cores:4}
	// 3: {Num:3 Start:18ms Percentage:7 Wall:{SweepTermination:100µs MarkAndScan:600µs MarkTermination:13µs} CPU:{SweepTermination:420µs MarkAssist:1ms MarkBackground:550µs MarkIdle:0s MarkTermination:55µs} Heap:{Before:4 After:4 Live:0 Goal:4} Cores:4}
}
```

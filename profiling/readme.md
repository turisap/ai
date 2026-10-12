Day 1 (Mon): Measure first, benchmarks

Learn: testing.B, b.ReportAllocs() / -benchmem, b.ResetTimer(), why results need a sink variable so the compiler can’t eliminate the work (or use b.Loop() on Go 1.24+), and benchstat for comparing runs.
Do:

Pick the hot function and write a benchmark with realistic inputs.
Run go test -bench=. -benchmem -count=10 > base.txt.
Run benchstat base.txt and check the variance.

Done when: you have a stable baseline (ns/op, B/op, allocs/op) saved to a file.
Read: the Go blog’s “Profiling Go Programs” and the first section of Dave Cheney’s High Performance Go workshop.
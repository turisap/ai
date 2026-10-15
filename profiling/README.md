# taskperf: a deliberately slow Go playground

Stdlib-only. Every function in `slow.go` has flaws (CPU, memory, contention, or a leak).
Spoilers are in `ANSWERS.md`: try to find them with the tools first.

Rules: keep `go test -run Test` green while you optimize (except the leak test, which
is *supposed* to fail until you fix the leak), and never merge a change without a benchstat diff.

Setup once: `go install golang.org/x/perf/cmd/benchstat@latest`

## Day 1: benchmarks and a baseline
```bash
go test -run='^$' -bench=. -benchmem -count=10 | tee base.txt
benchstat base.txt            # check the variance (± %) before trusting anything
```
`-run='^$'` skips the tests so only benchmarks run. After a change:
```bash
go test -run='^$' -bench=. -benchmem -count=10 | tee new.txt
benchstat base.txt new.txt
```
Task: write down your own guess of the top 3 offenders *before* profiling.

## Day 2: pprof
Offline (from benchmarks):
```bash
go test -run='^$' -bench=EncodeOutbox -benchtime=5s -cpuprofile cpu.out -memprofile mem.out
go tool pprof -http=:8080 cpu.out                       # flame graph, top, source
go tool pprof -sample_index=alloc_space -http=:8081 mem.out
go tool pprof -sample_index=alloc_objects -top mem.out
```
Contention (compare `-cpu=1,4,8` first):
```bash
go test -run='^$' -bench=CountersParallel -cpu=1,4,8 -mutexprofile mutex.out -blockprofile block.out
go tool pprof -http=:8082 mutex.out
```
Live service (pprof is on `localhost:6060`, app on `:8080`):
```bash
go run ./cmd/server -bg 4 &
curl -s localhost:8080/stats
go tool pprof -http=:8083 'http://localhost:6060/debug/pprof/profile?seconds=30'   # CPU
for i in $(seq 200); do curl -s localhost:8080/leak >/dev/null; done
curl -s 'localhost:6060/debug/pprof/goroutine?debug=1' | head -20                   # leak
for i in $(seq 100); do curl -s localhost:8080/trace >/dev/null; done
go tool pprof -sample_index=inuse_space -http=:8084 http://localhost:6060/debug/pprof/heap
```
Questions: which function is top by *flat* vs *cum*? Why does the heap profile for
`/trace` point where it does? What does the goroutine profile say after `/leak`?

## Day 3: escape analysis and allocations
```bash
go build -gcflags='-m' . 2>&1 | grep -E 'escapes|moved to heap'
go build -gcflags='-m -m' . 2>&1 | less      # also shows *why*
```
Map each "escapes to heap" line to the alloc counts from Day 1 (`allocs/op`) and to
`go tool pprof -list 'BuildFanOutKeys' mem.out`. Fix, re-run, benchstat.

## Day 4 (GC): run the live server under load
```bash
GODEBUG=gctrace=1 GOGC=100 go run ./cmd/server -bg 4 2>&1 | grep '^gc ' | tail
GOGC=400 ...    # then GOMEMLIMIT=256MiB GOGC=off ...
```

## Day 5: strings/[]byte, then a final before/after note.

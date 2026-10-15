# ANSWERS (spoilers)

Baseline from one run in a Linux container (Go 1.22); your numbers will differ.

| Benchmark | ns/op | B/op | allocs/op |
|---|---|---|---|
| BuildFanOutKeys (250 keys) | 60,215 | 15,424 | 709 |
| NormalizeStatus | 1,801 | 992 | 17 |
| CountByStatus (1000 tasks) | 1,526,212 | 962,254 | 15,449 |
| EncodeOutbox (200 tasks) | 1,746,237 | 8,502,961 | 5,801 |
| DedupeTags (5000 tasks) | 12,125,281 | 18,800 | 10 |
| Summarize (1000 tasks) | 263,102 | 116,461 | 3,756 |
| ShardFor | 22 | 0 | 0 |
| CountersParallel | 1,934 | 997 | 17 |

## Flaws and fixes

**1. NormalizeStatus / CountByStatus (CPU + allocs).** `regexp.MustCompile` runs on every call, so compile
cost dominates the CPU profile (`regexp/syntax.Parse`, `regexp.compile`). Fix: hoist to a package-level
`var re = regexp.MustCompile(...)`; then consider dropping regexp entirely (a hand-written loop, or a
`switch` over the known statuses with a fast path when the input is already normalized).

**2. EncodeOutbox (memory, the biggest B/op).** `payload += ...` is quadratic: every append copies the whole
string, hence ~8.5 MB/op for 200 events. Also a `map[string]interface{}` per event (boxing, key sorting in
`json.Marshal`) and `string(b)` / `[]byte(...)` conversions. Fixes, in order: `strings.Builder` or a single
`bytes.Buffer` with `Grow`; marshal a typed struct instead of a map; ultimately encode straight into one
buffer with `json.NewEncoder` or `AppendFormat`/`strconv.Append*`. Keep `TestEncodeOutbox` green.

**3. BuildFanOutKeys (allocs).** `fmt.Sprintf("%d", ...)` boxes its argument (`s escapes to heap` /
`userID escapes to heap` in `-m` output); string `+` makes intermediate strings; `keys` grows without
capacity. Fixes: `make([]string, 0, len(storeIDs)*len(statuses))`, compute the user-ID string once outside
the loops, lowercase each status once, build each key with `strconv.AppendInt` into a small buffer.

**4. DedupeTags (CPU).** O(n*u) linear scan with ~500 unique tags: pure CPU, only 10 allocs. Fix: a
`map[string]struct{}` (or sort + compact). In the CPU profile the time sits in the inner loop of `DedupeTags`
itself, so there is no allocation story here, which is why this one is invisible in `-benchmem` but obvious in pprof.

**5. Summarize (allocs).** `fmt.Sprintf` boxes `id`, `status`, `title` (see the `-m` lines for `v.id`, `v.status`,
`v.title`); `out` is not preallocated. Note `&taskView{}` in `Summarize` is reported as *not* escaping
(inlined and devirtualized), a good reminder that the interface is not the only cost. Fix: `make([]string, 0, len(tasks))`
and build each line with `strconv.AppendInt` + string concatenation into one buffer.

**6. Counters.IncTask (contention + CPU).** The lock is held while doing `fmt.Sprintf` and
`NormalizeStatus` (which compiles a regexp). Check how ns/op behaves with `-cpu=1,4,8` and what the mutex
profile blames. Fixes: do all the work before `Lock()`; then reduce contention further (sharded maps by
key hash, or `atomic.Int64` per pre-registered key).

**7. RememberTrace (memory leak by substring retention).** `id` is a substring of the log line, so the map key
keeps the whole ~2 KB line alive. Hit `/trace` repeatedly and watch `inuse_space`: the heap profile attributes
the bytes to the *allocation site* of the line (`GenLogLine` / `strings.Repeat`), not to `RememberTrace`, which is
the point of this exercise. Fix: `id = strings.Clone(id)`. A second bug: the map is unbounded. Add
eviction (TTL or a size cap) and discuss which you'd pick.

**8. FetchWithTimeout (goroutine leak).** The result channel is unbuffered, so after a timeout the worker
goroutine blocks forever on `ch <- ...`. `/leak` plus the goroutine profile shows N goroutines parked in
`chan send`. Fixes: `make(chan string, 1)`; better, pass a `context.Context` to `slowCall` so the work
itself is cancelled; use `time.NewTimer` and `Stop` it instead of `time.After` in hot loops.
`TestFetchWithTimeoutNoLeak` should go green.

**9. ShardFor (the control).** It looks wasteful (`fnv.New64a()`, `[]byte(key)`) but already shows 0 allocs/op
because the compiler inlines and stack-allocates both. 22 ns/op. Lesson: confirm with the profile before
"fixing" it. An inlined hand-rolled FNV loop may shave a few ns, but only matters if it's actually hot.

## Suggested optimization order
Biggest win per effort: (2) EncodeOutbox, (1) regexp hoist, (6) lock scope, (4) map dedupe, then (3)/(5),
then the leaks (7, 8). Write down the benchstat diff for each one.

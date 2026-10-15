// Server with a pprof endpoint on an internal port (:6060) and a few
// endpoints that exercise the flawed code paths from the taskperf package.
//
//	go run ./cmd/server -bg 4
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // registers /debug/pprof/* on http.DefaultServeMux
	"runtime"
	"sync/atomic"
	"time"

	"taskperf"
)

func main() {
	bg := flag.Int("bg", 0, "background load workers")
	flag.Parse()

	// Off by default; needed for /debug/pprof/mutex and /debug/pprof/block.
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(1000)

	// pprof lives on DefaultServeMux, bound to localhost only.
	go func() { log.Println(http.ListenAndServe("localhost:6060", nil)) }()

	tasks := taskperf.GenTasks(200, 1)
	stores := make([]int, 50)
	for i := range stores {
		stores[i] = i
	}
	sts := []string{"new", "in_progress", "blocked", "done", "cancelled"}
	counters := taskperf.NewCounters()
	var traceSeq atomic.Int64

	work := func() int {
		b, _ := taskperf.EncodeOutbox("batch", tasks)
		keys := taskperf.BuildFanOutKeys(stores, 42, sts)
		counts := taskperf.CountByStatus(tasks)
		_ = taskperf.DedupeTags(tasks)
		_ = taskperf.Summarize(tasks)
		for _, t := range tasks[:20] {
			counters.IncTask(t)
		}
		return len(b) + len(keys) + len(counts)
	}

	for i := 0; i < *bg; i++ {
		go func() {
			for {
				work()
			}
		}()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, work())
	})
	// Leaks one goroutine per call (see FetchWithTimeout).
	mux.HandleFunc("/leak", func(w http.ResponseWriter, r *http.Request) {
		_, ok := taskperf.FetchWithTimeout(10*time.Millisecond, func() string {
			time.Sleep(time.Hour)
			return "late"
		})
		fmt.Fprintln(w, "ok:", ok)
	})
	// Grows memory: each call remembers 1000 new trace ids.
	mux.HandleFunc("/trace", func(w http.ResponseWriter, r *http.Request) {
		for i := 0; i < 1000; i++ {
			taskperf.RememberTrace(taskperf.GenLogLine(int(traceSeq.Add(1))))
		}
		fmt.Fprintln(w, "seq:", traceSeq.Load())
	})
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Fprintf(w, "goroutines=%d heap_inuse_mb=%d num_gc=%d\n",
			runtime.NumGoroutine(), m.HeapInuse>>20, m.NumGC)
	})
	log.Println("app on :8080, pprof on localhost:6060")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

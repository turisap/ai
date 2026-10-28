package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // registers /debug/pprof/* on http.DefaultServeMux
	"runtime"

	"taskperf"
)

func makeTasks(n int) []taskperf.Task {
	tags := []string{"urgent", "backend", "billing", "infra", "bug", "feature", "ops", "security"}
	out := make([]taskperf.Task, n)
	for i := range out {
		out[i] = taskperf.Task{
			ID:     int64(i + 1),
			Title:  fmt.Sprintf("Task #%d: do the thing", i+1),
			Status: "in_progress",
			Tags:   []string{tags[i%8], tags[(i+3)%8], tags[(i+5)%8]},
		}
	}
	return out
}

func main() {
	bg := flag.Int("bg", 2, "background workers calling Process in a loop")
	flag.Parse()

	// pprof lives on http.DefaultServeMux; serve it on a localhost-only port,
	// separate from the app port. Never expose it publicly.
	go func() { log.Println(http.ListenAndServe("localhost:6060", nil)) }()

	tasks := makeTasks(200)
	for i := 0; i < *bg; i++ {
		go func() {
			for {
				taskperf.Process(tasks)
			}
		}()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, taskperf.Process(tasks))
	})
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Fprintf(w, "goroutines=%d heap_inuse_kb=%d num_gc=%d\n",
			runtime.NumGoroutine(), m.HeapInuse>>10, m.NumGC)
	})
	log.Println("app on :8080, pprof on localhost:6060")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

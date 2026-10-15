package taskperf

import (
	"encoding/json"
	"reflect"
	"runtime"
	"sort"
	"testing"
	"time"
)

// ---------- correctness tests: keep these green while you optimize ----------

func TestBuildFanOutKeys(t *testing.T) {
	got := BuildFanOutKeys([]int{1, 2}, 7, []string{"New", "DONE"})
	want := []string{"tasks:1:7:new", "tasks:1:7:done", "tasks:2:7:new", "tasks:2:7:done"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestNormalizeStatus(t *testing.T) {
	cases := map[string]string{
		" In Progress ": "in_progress",
		"DONE":          "done",
		"blocked":       "blocked",
		"a--b":          "a_b",
	}
	for in, want := range cases {
		if got := NormalizeStatus(in); got != want {
			t.Errorf("NormalizeStatus(%q)=%q want %q", in, got, want)
		}
	}
}

func TestCountByStatus(t *testing.T) {
	tasks := GenTasks(1000, 1)
	counts := CountByStatus(tasks)
	total := 0
	for k, v := range counts {
		total += v
		found := false
		for _, s := range statuses {
			found = found || s == k
		}
		if !found {
			t.Errorf("unexpected status key %q", k)
		}
	}
	if total != 1000 {
		t.Fatalf("total=%d", total)
	}
}

func TestEncodeOutbox(t *testing.T) {
	tasks := GenTasks(50, 1)
	b, err := EncodeOutbox("b-1", tasks)
	if err != nil {
		t.Fatal(err)
	}
	var out struct {
		BatchID string                   `json:"batch_id"`
		Events  []map[string]interface{} `json:"events"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out.BatchID != "b-1" || len(out.Events) != 50 {
		t.Fatalf("batch=%q events=%d", out.BatchID, len(out.Events))
	}
	e := out.Events[0]
	if e["id"] != float64(1) || e["type"] != "task.updated" || e["title"] != tasks[0].Title {
		t.Fatalf("bad first event: %v", e)
	}
	if e["created_at"] != tasks[0].CreatedAt.Format(time.RFC3339) {
		t.Fatalf("bad created_at: %v", e["created_at"])
	}
}

func TestDedupeTags(t *testing.T) {
	tasks := GenTasks(2000, 1)
	got := DedupeTags(tasks)
	set := map[string]struct{}{}
	for _, tk := range tasks {
		for _, tag := range tk.Tags {
			set[tag] = struct{}{}
		}
	}
	want := make([]string, 0, len(set))
	for k := range set {
		want = append(want, k)
	}
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %d tags want %d", len(got), len(want))
	}
}

func TestSummarize(t *testing.T) {
	tasks := GenTasks(10, 1)
	got := Summarize(tasks)
	if len(got) != 10 {
		t.Fatalf("len=%d", len(got))
	}
	want := "1|" + tasks[0].Status + "|" + tasks[0].Title
	if got[0] != want {
		t.Fatalf("got %q want %q", got[0], want)
	}
}

func TestShardFor(t *testing.T) {
	golden := map[string]int{
		"tasks:1:7:new": 15,
		"user:42":       2,
	}
	for k, want := range golden {
		if got := ShardFor(k, 64); got != want {
			t.Errorf("ShardFor(%q)=%d want %d", k, got, want)
		}
	}
}

func TestRememberTrace(t *testing.T) {
	line := GenLogLine(3)
	id := RememberTrace(line)
	if len(id) != 16 {
		t.Fatalf("id=%q", id)
	}
}

func TestFetchWithTimeout(t *testing.T) {
	v, ok := FetchWithTimeout(100*time.Millisecond, func() string { return "x" })
	if !ok || v != "x" {
		t.Fatalf("fast call: %q %v", v, ok)
	}
	_, ok = FetchWithTimeout(time.Millisecond, func() string { time.Sleep(50 * time.Millisecond); return "y" })
	if ok {
		t.Fatal("expected timeout")
	}
}

// Expected to FAIL until you fix the goroutine leak (Day 2: goroutine profile).
func TestFetchWithTimeoutNoLeak(t *testing.T) {
	before := runtime.NumGoroutine()
	for i := 0; i < 100; i++ {
		FetchWithTimeout(time.Millisecond, func() string { time.Sleep(20 * time.Millisecond); return "y" })
	}
	time.Sleep(200 * time.Millisecond)
	if after := runtime.NumGoroutine(); after > before+2 {
		t.Fatalf("goroutine leak: before=%d after=%d", before, after)
	}
}

// ---------- Day 1 benchmarks ----------
// Sinks stop the compiler from optimizing the work away.
var (
	sinkStrs  []string
	sinkBytes []byte
	sinkMap   map[string]int
	sinkInt   int
	sinkStr   string
)

func BenchmarkBuildFanOutKeys(b *testing.B) {
	stores := make([]int, 50)
	for i := range stores {
		stores[i] = i
	}
	sts := []string{"new", "in_progress", "blocked", "done", "cancelled"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStrs = BuildFanOutKeys(stores, 42, sts)
	}
}

func BenchmarkNormalizeStatus(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sinkStr = NormalizeStatus("  In_Progress ")
	}
}

func BenchmarkCountByStatus(b *testing.B) {
	tasks := GenTasks(1000, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkMap = CountByStatus(tasks)
	}
}

func BenchmarkEncodeOutbox(b *testing.B) {
	tasks := GenTasks(200, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkBytes, _ = EncodeOutbox("batch-1", tasks)
	}
}

func BenchmarkDedupeTags(b *testing.B) {
	tasks := GenTasks(5000, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStrs = DedupeTags(tasks)
	}
}

func BenchmarkSummarize(b *testing.B) {
	tasks := GenTasks(1000, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkStrs = Summarize(tasks)
	}
}

func BenchmarkShardFor(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sinkInt = ShardFor("tasks:12:4242:in_progress", 64)
	}
}

// Day 2: run with -mutexprofile / -blockprofile and compare -cpu=1,4,8.
func BenchmarkCountersParallel(b *testing.B) {
	c := NewCounters()
	tasks := GenTasks(1000, 1)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			c.IncTask(tasks[i%len(tasks)])
			i++
		}
	})
}

package taskperf

import (
	"fmt"
	"testing"
)

var sinkInt int

func makeTasks(n int) []Task {
	tags := []string{"urgent", "backend", "billing", "infra", "bug", "feature", "ops", "security"}
	out := make([]Task, n)
	for i := range out {
		out[i] = Task{
			ID:     int64(i + 1),
			Title:  fmt.Sprintf("Task #%d: do the thing", i+1),
			Status: "in_progress",
			Tags:   []string{tags[i%8], tags[(i+3)%8], tags[(i+5)%8]},
		}
	}
	return out
}

func TestRender(t *testing.T) {
	got := render(makeTasks(1)[0])
	want := "1|IN_PROGRESS|Task #1: do the thing|feature,infra,urgent"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestProcess(t *testing.T) {
	if got := Process(makeTasks(10)); got != 4601 {
		t.Fatalf("got %d", got)
	}
}

func BenchmarkProcess(b *testing.B) {
	tasks := makeTasks(1000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkInt = Process(tasks)
	}
}

package taskperf

import (
	"reflect"
	"testing"
)

func TestBuildFanOutKeys(t *testing.T) {
	got := BuildFanOutKeys([]int{1, 2}, 7, []string{"New", "DONE"})
	want := []string{"tasks:1:7:new", "tasks:1:7:done", "tasks:2:7:new", "tasks:2:7:done"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

// sink keeps the compiler from optimizing the call away.
var sinkStrs []string

func BenchmarkBuildFanOutKeys(b *testing.B) {
	stores := make([]int, 50)
	for i := range stores {
		stores[i] = i
	}
	sts := []string{"new", "in_progress", "blocked", "done", "cancelled"}

	b.ReportAllocs()
	b.ResetTimer() // exclude the setup above from the measurement
	for i := 0; i < b.N; i++ {
		sinkStrs = BuildFanOutKeys(stores, 42, sts)
	}
}

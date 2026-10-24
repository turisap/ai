package taskperf

import (
	"reflect"
	"strconv"
	"strings"
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

func TestBuildFanOutKeysLargeIDs(t *testing.T) {
	stores := []int{0, 5, 10, 99, 100, 12345, 987654321}
	sts := []string{"New", "IN_PROGRESS", "x"}
	var want []string
	for _, s := range stores {
		for _, st := range sts {
			want = append(want, "tasks:"+strconv.Itoa(s)+":123456789:"+strings.ToLower(st))
		}
	}
	got := BuildFanOutKeys(stores, 123456789, sts)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v\nwant %v", got, want)
	}
}

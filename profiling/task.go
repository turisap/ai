package taskperf

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type Task struct {
	ID        int64     `json:"id"`
	StoreID   int       `json:"store_id"`
	UserID    int64     `json:"user_id"`
	Status    string    `json:"status"`
	Title     string    `json:"title"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
}

var statuses = []string{"new", "in_progress", "blocked", "done", "cancelled"}

// tagPool is a universe of 500 distinct tags.
var tagPool = func() []string {
	p := make([]string, 500)
	for i := range p {
		p[i] = fmt.Sprintf("tag-%03d", i)
	}
	return p
}()

// GenTasks returns deterministic synthetic tasks. Statuses are deliberately
// messy (mixed case, padding), like data coming from an external source.
func GenTasks(n int, seed int64) []Task {
	r := rand.New(rand.NewSource(seed))
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	out := make([]Task, n)
	for i := range out {
		tags := make([]string, 1+r.Intn(5))
		for j := range tags {
			tags[j] = tagPool[r.Intn(len(tagPool))]
		}
		st := statuses[r.Intn(len(statuses))]
		switch r.Intn(4) {
		case 0:
			st = strings.ToUpper(st)
		case 1:
			st = " " + st + " "
		}
		out[i] = Task{
			ID:        int64(i + 1),
			StoreID:   r.Intn(50),
			UserID:    int64(r.Intn(10_000)),
			Status:    st,
			Title:     fmt.Sprintf("Task #%d: do the thing", i+1),
			Tags:      tags,
			CreatedAt: base.Add(time.Duration(r.Intn(86400*90)) * time.Second),
		}
	}
	return out
}

// GenLogLine returns a ~2KB log line containing a unique trace_id.
func GenLogLine(i int) string {
	return fmt.Sprintf("ts=2026-10-21T10:00:00Z level=info trace_id=%016x msg=%q",
		uint64(i)*2654435761, strings.Repeat("payload ", 250))
}

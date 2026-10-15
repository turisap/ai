package taskperf

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// BuildFanOutKeys builds Redis keys: tasks:{store}:{user}:{status}
func BuildFanOutKeys(storeIDs []int, userID int64, statuses []string) []string {
	var keys []string
	for _, s := range storeIDs {
		for _, st := range statuses {
			key := "tasks:" + fmt.Sprintf("%d", s) + ":" + fmt.Sprintf("%d", userID) + ":" + strings.ToLower(st)
			keys = append(keys, key)
		}
	}
	return keys
}

// NormalizeStatus lowercases, trims and replaces non-alphanumerics with "_".
func NormalizeStatus(s string) string {
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = strings.ToLower(strings.TrimSpace(s))
	return re.ReplaceAllString(s, "_")
}

// CountByStatus counts tasks per normalized status.
func CountByStatus(tasks []Task) map[string]int {
	counts := map[string]int{}
	for _, t := range tasks {
		counts[NormalizeStatus(t.Status)]++
	}
	return counts
}

// EncodeOutbox serializes tasks into {"batch_id":...,"events":[...]}.
func EncodeOutbox(batchID string, tasks []Task) ([]byte, error) {
	payload := ""
	for i, t := range tasks {
		ev := map[string]interface{}{
			"type":       "task.updated",
			"id":         t.ID,
			"store_id":   t.StoreID,
			"user_id":    t.UserID,
			"status":     t.Status,
			"title":      t.Title,
			"tags":       t.Tags,
			"created_at": t.CreatedAt.Format(time.RFC3339),
		}
		b, err := json.Marshal(ev)
		if err != nil {
			return nil, err
		}
		if i > 0 {
			payload += ","
		}
		payload += string(b)
	}
	return []byte(`{"batch_id":"` + batchID + `","events":[` + payload + `]}`), nil
}

// DedupeTags returns the sorted set of unique tags across all tasks.
func DedupeTags(tasks []Task) []string {
	var uniq []string
	for _, t := range tasks {
		for _, tag := range t.Tags {
			found := false
			for _, u := range uniq {
				if u == tag {
					found = true
					break
				}
			}
			if !found {
				uniq = append(uniq, tag)
			}
		}
	}
	sort.Strings(uniq)
	return uniq
}

type Describer interface{ Describe() string }

type taskView struct {
	id            int64
	status, title string
}

func (v *taskView) Describe() string {
	return fmt.Sprintf("%d|%s|%s", v.id, v.status, v.title)
}

func newView(t Task) *taskView { return &taskView{t.ID, t.Status, t.Title} }

// Summarize returns one "id|status|title" line per task.
func Summarize(tasks []Task) []string {
	var out []string
	for _, t := range tasks {
		var d Describer = newView(t)
		out = append(out, d.Describe())
	}
	return out
}

// ShardFor maps a key to a shard with FNV-1a.
func ShardFor(key string, shards int) int {
	h := fnv.New64a()
	h.Write([]byte(key))
	return int(h.Sum64() % uint64(shards))
}

// Counters counts tasks per (store,status).
type Counters struct {
	mu sync.Mutex
	m  map[string]int64
}

func NewCounters() *Counters { return &Counters{m: map[string]int64{}} }

func (c *Counters) IncTask(t Task) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := fmt.Sprintf("store:%d:%s", t.StoreID, NormalizeStatus(t.Status))
	c.m[key]++
}

func (c *Counters) Total() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	var n int64
	for _, v := range c.m {
		n += v
	}
	return n
}

var (
	traceMu    sync.Mutex
	seenTraces = map[string]time.Time{}
)

// RememberTrace extracts the trace id from a log line and records when we saw it.
func RememberTrace(line string) string {
	i := strings.Index(line, "trace_id=")
	if i < 0 || len(line) < i+9+16 {
		return ""
	}
	id := line[i+9 : i+9+16]
	traceMu.Lock()
	seenTraces[id] = time.Now()
	traceMu.Unlock()
	return id
}

// FetchWithTimeout runs slowCall and gives up after timeout.
func FetchWithTimeout(timeout time.Duration, slowCall func() string) (string, bool) {
	ch := make(chan string)
	go func() { ch <- slowCall() }()
	select {
	case v := <-ch:
		return v, true
	case <-time.After(timeout):
		return "", false
	}
}

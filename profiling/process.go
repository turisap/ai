package taskperf

import (
	"sort"
	"strconv"
	"strings"
)

type Task struct {
	ID     int64
	Title  string
	Status string
	Tags   []string
}

// Process scores and renders every task and returns a checksum.
func Process(tasks []Task) int {
	total := 0
	for _, t := range tasks {
		total += score(t)
		total += len(render(t))
	}
	return total
}

// score is a CPU-bound leaf.
func score(t Task) int {
	h := uint32(2166136261)
	for round := 0; round < 40; round++ {
		for i := 0; i < len(t.Title); i++ {
			h ^= uint32(t.Title[i])
			h *= 16777619
		}
	}
	return int(h % 1000)
}

// render does little itself but calls several other things.
// func render(t Task) string {
// 	tags := make([]string, len(t.Tags))
// 	copy(tags, t.Tags)
// 	sort.Strings(tags)
// 	return fmt.Sprintf("%d|%s|%s|%s", t.ID, strings.ToUpper(t.Status), t.Title, strings.Join(tags, ","))
// }

func render(t Task) string {
	var tagsBuf [8]string
	tags := append(tagsBuf[:0], t.Tags...) // copy, so the caller's slice isn't reordered
	sort.Strings(tags)

	status := strings.ToUpper(t.Status)
	size := 20 + len(status) + len(t.Title) + 3
	for _, tg := range tags {
		size += len(tg) + 1
	}

	var sb strings.Builder
	sb.Grow(size)
	var num [20]byte
	sb.Write(strconv.AppendInt(num[:0], t.ID, 10))
	sb.WriteByte('|')
	sb.WriteString(status)
	sb.WriteByte('|')
	sb.WriteString(t.Title)
	sb.WriteByte('|')
	for i, tg := range tags {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(tg)
	}
	return sb.String()
}

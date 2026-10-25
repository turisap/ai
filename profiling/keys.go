package taskperf

import (
	"strconv"
	"strings"
)

// BuildFanOutKeys builds Redis keys: tasks:{store}:{user}:{status}
func BuildFanOutKeys(storeIDs []int, userID int64, statuses []string) []string {
	// the ":{user}:{status}" suffix is the same for every store: build it once
	suffixes := make([][]byte, len(statuses))
	perStore := len(statuses) * 10 // "tasks:" + a few digits, per key
	for i, st := range statuses {
		b := make([]byte, 0, 32)
		b = append(b, ':')
		b = strconv.AppendInt(b, userID, 10)
		b = append(b, ':')
		b = append(b, strings.ToLower(st)...)
		suffixes[i] = b
		perStore += len(b)
	}

	// all key text goes into ONE buffer; each key is a substring of it
	var sb strings.Builder
	sb.Grow(len(storeIDs) * perStore)
	var digits [20]byte
	keys := make([]string, 0, len(storeIDs)*len(statuses))
	for _, s := range storeIDs {
		id := strconv.AppendInt(digits[:0], int64(s), 10)
		for _, suf := range suffixes {
			// @FIXME this is bad.
			// sb.Reset()
			start := sb.Len() // this is good - writing to the buffer in append-only
			sb.WriteString("tasks:")
			sb.Write(id)
			sb.Write(suf)
			// keys = append(keys, sb.String()[:])
			keys = append(keys, sb.String()[start:])
		}
	}
	return keys
}

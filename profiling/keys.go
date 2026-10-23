package taskperf

import (
	"strconv"
	"strings"
)

// BuildFanOutKeys builds Redis keys: tasks:{store}:{user}:{status}
func BuildFanOutKeys(storeIDs []int, userID int64, statuses []string) []string {
	// the ":{user}:{status}" suffix is the same for every store: build it once
	suffixes := make([][]byte, len(statuses))
	for i, st := range statuses {
		b := make([]byte, 0, 32)
		b = append(b, ':')
		b = strconv.AppendInt(b, userID, 10)
		b = append(b, ':')
		b = append(b, strings.ToLower(st)...)
		suffixes[i] = b
	}

	keys := make([]string, 0, len(storeIDs)*len(statuses))
	buf := make([]byte, 0, 64)
	for _, s := range storeIDs {
		buf = append(buf[:0], "tasks:"...)
		buf = strconv.AppendInt(buf, int64(s), 10)
		prefixLen := len(buf)
		for _, suf := range suffixes {
			buf = append(buf[:prefixLen], suf...)
			keys = append(keys, string(buf))
		}
	}
	return keys
}

package taskperf

import (
	"strconv"
	"strings"
)

// BuildFanOutKeys builds Redis keys: tasks:{store}:{user}:{status}
func BuildFanOutKeys(storeIDs []int, userID int64, statuses []string) []string {
	keys := make([]string, 0, len(storeIDs)*len(statuses))
	buf := make([]byte, 0, 64)
	for _, s := range storeIDs {
		for _, st := range statuses {
			buf = append(buf[:0], "tasks:"...)
			buf = strconv.AppendInt(buf, int64(s), 10)
			buf = append(buf, ':')
			buf = strconv.AppendInt(buf, userID, 10)
			buf = append(buf, ':')
			buf = append(buf, strings.ToLower(st)...)
			keys = append(keys, string(buf))
		}
	}
	return keys
}

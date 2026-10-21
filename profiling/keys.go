package taskperf

import (
	"fmt"
	"strings"
)

// BuildFanOutKeys builds Redis keys: tasks:{store}:{user}:{status}
func BuildFanOutKeys(storeIDs []int, userID int64, statuses []string) []string {
	keys := make([]string, 0, len(storeIDs)*len(statuses))
	for _, s := range storeIDs {
		for _, st := range statuses {
			key := "tasks:" + fmt.Sprintf("%d", s) + ":" + fmt.Sprintf("%d", userID) + ":" + strings.ToLower(st)
			keys = append(keys, key)
		}
	}
	return keys
}

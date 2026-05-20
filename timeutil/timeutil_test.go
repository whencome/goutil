package timeutil

import (
	"testing"
	"time"
)

// TestSplitTimeRange 测试日期拆分
func TestSplitTimeRange(t *testing.T) {
	startTime := Unix("2023-01-01")
	endTime := Unix("2023-12-07")
	cutFunc := func(startTime, endTime time.Time) (TimeRange, bool) {
		hasNext := true
		end := startTime.AddDate(0, 0, 16)
		if startTime.Month() != end.Month() {
			end = time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, end.Location())
		}
		if end.After(endTime) || end.Equal(endTime) {
			end = endTime
			hasNext = false
		}
		return TimeRange{StartTime: startTime, EndTime: end}, hasNext
	}
	timeRanges := SplitTimeRange(startTime, endTime, cutFunc)
	if len(timeRanges) != 23 {
		t.Logf("expect time range count = 23, but got %d", len(timeRanges))
		t.Fail()
	}
	for i, tr := range timeRanges {
		t.Logf("%d: %s - %s\n", i, tr.StartTime.Format(time.DateTime), tr.EndTime.Format(time.DateTime))
	}
}

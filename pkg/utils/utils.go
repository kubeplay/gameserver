package utils

import "time"

func RoundTime(d, r time.Duration) time.Duration {
	if r <= 0 {
		return d
	}
	neg := d < 0
	if neg {
		d = -d
	}
	if m := d % r; m+m < r {
		d = d - m
	} else {
		d = d + r - m
	}
	if neg {
		return -d
	}
	return d
}

func GetDeltaDuration(startTime, endTime string) time.Duration {
	start, _ := time.Parse(time.RFC3339, startTime)
	end, _ := time.Parse(time.RFC3339, endTime)
	delta := end.Sub(start)
	var d time.Duration
	if endTime != "" {
		d = RoundTime(delta, time.Second)
	} else {
		d = RoundTime(time.Since(start), time.Second)
	}
	return d
}

package uptime

import (
	"strconv"
	"strings"
	"time"
)

// Clock interface makes it easier to test
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (r *realClock) Now() time.Time {
	return time.Now()
}

func parseLinuxUptime(uptime string) (float64, error) {
	parts := strings.Split(string(uptime), " ")
	secondsStr := strings.TrimSpace(parts[0])
	return strconv.ParseFloat(secondsStr, 64)
}
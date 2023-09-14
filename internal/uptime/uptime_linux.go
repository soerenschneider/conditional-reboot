package uptime

import (
	"fmt"
	"math"
	"os"
	"time"
)

var (
	rawUptimeImpl UptimeSource = &LinuxUptime{}
	clock         Clock        = &realClock{}
)

type UptimeSource interface {
	RawUptime() (float64, error)
}

func Uptime() (time.Duration, error) {
	seconds, err := rawUptimeImpl.RawUptime()
	if err != nil {
		return time.Duration(0), fmt.Errorf("could not read raw uptime: %w", err)
	}

	return time.Second * time.Duration(seconds), nil
}

type LinuxUptime struct {
}

func (p *LinuxUptime) RawUptime() (float64, error) {
	uptime, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return math.MaxFloat64, err
	}
	return parseLinuxUptime(string(uptime))
}

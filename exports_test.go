package timertask

import (
	"testing"
	"time"
)

func MockSleepInterval(t *testing.T, duration time.Duration) {
	old := sleepDuration
	sleepDuration = duration
	t.Cleanup(func() { sleepDuration = old })
}

func MockUnixTimeNow(t *testing.T, timeFunc func() int64) {
	old := unixTimeNow
	unixTimeNow = timeFunc
	t.Cleanup(func() { unixTimeNow = old })
}

func (m *Manager) WorkersByIntervals() map[int64]int64 {
	return m.workersByIntervals
}

func (m *Manager) Workers() map[int64]map[int64]taskConfig {
	return m.workers
}

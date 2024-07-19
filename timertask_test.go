package timertask_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/xconnio/timertask"
)

func TestSchedule(t *testing.T) {
	manager := timertask.NewManager()
	callback := func() error {
		return nil
	}
	manager.Schedule(1, 1*time.Second, callback)
	require.NotNil(t, manager.WorkersByIntervals()[1])
	require.Len(t, manager.Workers(), 1)
}

func TestCancel(t *testing.T) {
	manager := timertask.NewManager()
	callback := func() error {
		return nil
	}
	manager.Schedule(1, 1*time.Second, callback)
	manager.Cancel(1)
	_, exists := manager.WorkersByIntervals()[1]
	require.False(t, exists)
	require.Len(t, manager.Workers(), 0)
}

func TestReset(t *testing.T) {
	manager := timertask.NewManager()
	callback := func() error {
		return nil
	}
	manager.Schedule(1, 1*time.Second, callback)
	manager.Reset(1)

	require.NotNil(t, manager.WorkersByIntervals()[1])
	require.Len(t, manager.Workers(), 1)
}

func TestStart(t *testing.T) {
	timertask.MockSleepInterval(t, 1*time.Millisecond)
	initialTime := time.Now()
	timertask.MockUnixTimeNow(t, func() int64 {
		return initialTime.Unix()
	})

	manager := timertask.NewManager()
	callbackCalled := make(chan bool, 1)
	callback := func() error {
		callbackCalled <- true
		return nil
	}
	manager.Schedule(1, 10*time.Millisecond, callback)
	manager.Start()

	select {
	case <-callbackCalled:
	case <-time.After(20 * time.Millisecond):
		t.Fatal("Callback was not called in time")
	}
}

package timertask_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/xconnio/timertask"
)

func TestManager_Schedule(t *testing.T) {
	manager := timertask.NewManager()

	messages1 := 0
	manager.Schedule(1, 2*time.Second, func() error {
		messages1++
		return nil
	})

	messages2 := 0
	manager.Schedule(2, 1*time.Second, func() error {
		messages2++
		return nil
	})

	go manager.Start()

	// Wait for worker
	time.Sleep(3 * time.Second)

	require.NotZero(t, messages1)
	require.NotZero(t, messages2)

	// Cancel worker1 and check
	manager.Cancel(1)
	time.Sleep(3 * time.Second)

	require.Equal(t, messages1, 1)

	require.Greater(t, messages2, 2)
}

func TestManager_Reset(t *testing.T) {
	manager := timertask.NewManager()

	messages := 0
	manager.Schedule(1, 2*time.Second, func() error {
		messages++
		return nil
	})

	manager.Start()

	// Wait for first work
	time.Sleep(3 * time.Second)
	require.Equal(t, 1, messages)

	time.Sleep(900 * time.Millisecond)
	// Reset worker1 and check work
	manager.Reset(1)

	time.Sleep(1 * time.Second)
	require.Equal(t, 1, messages)

	// Wait for another work after reset
	time.Sleep(2 * time.Second)
	require.Equal(t, 2, messages)
}

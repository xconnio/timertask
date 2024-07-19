package workerpool_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/xconnio/workerpool"
)

func TestManager(t *testing.T) {
	manager := workerpool.NewWorkerPool()

	messages1 := 0
	manager.Add(1, 2*time.Second, func() error {
		messages1++
		return nil
	})

	messages2 := 0
	manager.Add(2, 1*time.Second, func() error {
		messages2++
		return nil
	})

	go manager.Start()

	// Wait for worker
	time.Sleep(3 * time.Second)

	require.NotZero(t, messages1)
	require.NotZero(t, messages2)

	// Remove worker1 and check
	manager.Remove(1)
	time.Sleep(3 * time.Second)

	require.Equal(t, messages1, 1)

	require.Greater(t, messages2, 2)
}

func TestManagerReset(t *testing.T) {
	manager := workerpool.NewWorkerPool()

	messages := 0
	manager.Add(1, 2*time.Second, func() error {
		messages++
		return nil
	})

	go manager.Start()

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

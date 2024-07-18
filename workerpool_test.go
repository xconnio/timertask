package workerpool_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/xconnio/workerpool"
)

type mockClient struct {
	messages []string
}

func (c *mockClient) Write(data []byte) error {
	c.messages = append(c.messages, string(data))
	return nil
}

func TestManager(t *testing.T) {
	manager := workerpool.NewPingManager()

	client1 := &mockClient{}
	client2 := &mockClient{}

	manager.Add(client1, 2*time.Second, func() []byte {
		return []byte("ping")
	})
	manager.Add(client2, 1*time.Second, func() []byte {
		return []byte("ping")
	})

	go manager.Start()

	// Wait for ping
	time.Sleep(3 * time.Second)

	require.NotEmpty(t, client1.messages)
	require.NotEmpty(t, client2.messages)

	// Remove client1 and check
	manager.Remove(client1)
	time.Sleep(3 * time.Second)

	require.Len(t, client1.messages, 1)

	require.Greater(t, len(client2.messages), 2)
}

func TestManagerReset(t *testing.T) {
	manager := workerpool.NewPingManager()

	client1 := &mockClient{}
	manager.Add(client1, 2*time.Second, func() []byte {
		return []byte("ping")
	})

	go manager.Start()

	// Wait for first ping
	time.Sleep(3 * time.Second)
	require.Len(t, client1.messages, 1)

	time.Sleep(1 * time.Second)
	// Reset client1 and check pings
	manager.Reset(client1)

	time.Sleep(1 * time.Second)
	require.Len(t, client1.messages, 1)

	// Wait for another ping after reset
	time.Sleep(2 * time.Second)
	require.Len(t, client1.messages, 2)
}

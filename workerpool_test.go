package workerpool_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/xconnio/workerpool"
)

type mockClient struct {
	interval time.Duration
	messages []string
}

func (c *mockClient) WriteMessage(data []byte) error {
	c.messages = append(c.messages, string(data))
	return nil
}

func (c *mockClient) PingInterval() time.Duration {
	return c.interval
}

func TestManager(t *testing.T) {
	manager := workerpool.NewPingManager()

	client1 := &mockClient{interval: 2 * time.Second}
	client2 := &mockClient{interval: 1 * time.Second}

	manager.Add(client1)
	manager.Add(client2)

	go manager.Start()

	// Wait for ping
	time.Sleep(3 * time.Second)

	require.NotEmpty(t, client1.messages)
	require.NotEmpty(t, client2.messages)

	// Remove client1 and check
	manager.Remove(client1)
	time.Sleep(3 * time.Second)

	require.Len(t, client1.messages, 2)

	require.Greater(t, len(client2.messages), 2)
}

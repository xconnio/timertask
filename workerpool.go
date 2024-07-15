package workerpool

import (
	"log"
	"sync"
	"time"
)

type Client interface {
	WriteMessage(data []byte) error
	PingInterval() time.Duration
}

type PingManager struct {
	clients       map[int64][]Client
	clientPingMap map[Client]int64
	mutex         sync.Mutex
}

// NewPingManager creates a new PingManager instance.
func NewPingManager() *PingManager {
	return &PingManager{
		clients:       make(map[int64][]Client),
		clientPingMap: make(map[Client]int64),
	}
}

// Add registers a client to be pinged.
func (m *PingManager) Add(client Client) {
	// Calculate the next ping time in seconds
	timeToPing := time.Now().Add(client.PingInterval()).Unix()

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.clients[timeToPing] = append(m.clients[timeToPing], client)
	m.clientPingMap[client] = timeToPing
}

// Remove unregisters a client.
func (m *PingManager) Remove(client Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	timeToPing, ok := m.clientPingMap[client]
	if ok {
		clients := m.clients[timeToPing]
		for i, c := range clients {
			if c == client {
				// Remove the client from the list
				m.clients[timeToPing] = append(clients[:i], clients[i+1:]...)
				break
			}
		}

		delete(m.clientPingMap, client)
	}
}

// Start begins the pinging process.
func (m *PingManager) Start() {
	go func() {
		for {
			nowSeconds := time.Now().Unix()

			m.mutex.Lock()
			clients, ok := m.clients[nowSeconds]
			if ok {
				for _, client := range clients {
					err := client.WriteMessage([]byte("ping"))
					if err != nil {
						log.Printf("error pinging client: %v", err)
						m.Remove(client)
						continue
					}

					// Reschedule the next ping for the client
					timeToPing := time.Now().Add(client.PingInterval()).Unix()
					m.clients[timeToPing] = append(m.clients[timeToPing], client)
					m.clientPingMap[client] = timeToPing

				}
				delete(m.clients, nowSeconds)
			}
			m.mutex.Unlock()

			time.Sleep(1 * time.Second)
		}
	}()
}

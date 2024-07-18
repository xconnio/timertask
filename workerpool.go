package workerpool

import (
	"log"
	"sync"
	"time"
)

type Client interface {
	WriteMessage(data []byte) error
}

type PingManager struct {
	clients       map[int64]map[Client]time.Duration
	clientPingMap map[Client]int64
	mutex         sync.Mutex
}

// NewPingManager creates a new PingManager instance.
func NewPingManager() *PingManager {
	return &PingManager{
		clients:       make(map[int64]map[Client]time.Duration),
		clientPingMap: make(map[Client]int64),
	}
}

// Add registers a client to be pinged.
func (m *PingManager) Add(client Client, pingInterval time.Duration) {
	// Calculate the next ping time in seconds
	timeToPing := time.Now().Add(pingInterval).Unix()

	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.clients[timeToPing] == nil {
		m.clients[timeToPing] = make(map[Client]time.Duration)
	}

	m.clients[timeToPing][client] = pingInterval
	m.clientPingMap[client] = timeToPing
}

// Remove unregisters a client.
func (m *PingManager) Remove(client Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	nextPingTime, ok := m.clientPingMap[client]
	if ok {
		delete(m.clients[nextPingTime], client)
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
				for client, interval := range clients {
					err := client.WriteMessage([]byte("ping"))
					if err != nil {
						log.Printf("error pinging client: %v", err)
						m.Remove(client)
						continue
					}

					// Reschedule the next ping for the client
					timeToPing := time.Now().Add(interval).Unix()
					if m.clients[timeToPing] == nil {
						m.clients[timeToPing] = make(map[Client]time.Duration)
					}
					m.clients[timeToPing][client] = interval
					m.clientPingMap[client] = timeToPing

				}
				delete(m.clients, nowSeconds)
			}
			m.mutex.Unlock()

			time.Sleep(1 * time.Second)
		}
	}()
}

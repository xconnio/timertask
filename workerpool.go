package workerpool

import (
	"log"
	"sync"
	"time"
)

type Client interface {
	Write(data []byte) error
}

type pingConfig struct {
	interval time.Duration
	dataFunc func() []byte
}

type PingManager struct {
	clients       map[int64]map[Client]pingConfig
	clientPingMap map[Client]int64
	mutex         sync.Mutex
}

// NewPingManager creates a new PingManager instance.
func NewPingManager() *PingManager {
	return &PingManager{
		clients:       make(map[int64]map[Client]pingConfig),
		clientPingMap: make(map[Client]int64),
	}
}

// Add registers a client to be pinged.
func (m *PingManager) Add(client Client, pingInterval time.Duration, dataFunc func() []byte) {
	// Calculate the next ping time in seconds
	timeToPing := time.Now().Add(pingInterval).Unix()

	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.clients[timeToPing] == nil {
		m.clients[timeToPing] = make(map[Client]pingConfig)
	}

	m.clients[timeToPing][client] = pingConfig{
		interval: pingInterval,
		dataFunc: dataFunc,
	}
	m.clientPingMap[client] = timeToPing
}

// Remove unregisters a client.
func (m *PingManager) Remove(client Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	nextPingTime, ok := m.clientPingMap[client]
	if ok {
		delete(m.clients[nextPingTime], client)
		if len(m.clients[nextPingTime]) == 0 {
			delete(m.clients, nextPingTime)
		}
		delete(m.clientPingMap, client)
	}
}

// Reset reschedules the ping for a specific client.
func (m *PingManager) Reset(client Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	nextPingTime, ok := m.clientPingMap[client]
	if ok {
		pingInfo := m.clients[nextPingTime][client]
		delete(m.clients[nextPingTime], client)
		if len(m.clients[nextPingTime]) == 0 {
			delete(m.clients, nextPingTime)
		}

		newPingTime := time.Now().Add(pingInfo.interval).Unix()
		if m.clients[newPingTime] == nil {
			m.clients[newPingTime] = make(map[Client]pingConfig)
		}
		m.clients[newPingTime][client] = pingInfo
		m.clientPingMap[client] = newPingTime
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
				for client, pingCfg := range clients {
					err := client.Write(pingCfg.dataFunc())
					if err != nil {
						log.Printf("error pinging client: %v", err)
						m.Remove(client)
						continue
					}

					// Reschedule the next ping for the client
					timeToPing := time.Now().Add(pingCfg.interval).Unix()
					if m.clients[timeToPing] == nil {
						m.clients[timeToPing] = make(map[Client]pingConfig)
					}
					m.clients[timeToPing][client] = pingCfg
					m.clientPingMap[client] = timeToPing

				}
				delete(m.clients, nowSeconds)
			}
			m.mutex.Unlock()

			time.Sleep(1 * time.Second)
		}
	}()
}

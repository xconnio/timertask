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

type PingManager interface {
	Register(client Client)
	Unregister(client Client)
	Start()
}

type Manager struct {
	clients map[Client]time.Time
	mutex   sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		clients: make(map[Client]time.Time),
	}
}

func (m *Manager) Register(client Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.clients[client] = time.Now()
}

func (m *Manager) Unregister(client Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.clients, client)
}

func (m *Manager) Start() {
	go func() {
		for {
			m.pingClients()
			time.Sleep(1 * time.Second)
		}
	}()
}

func (m *Manager) pingClients() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for client, lastActive := range m.clients {
		if time.Since(lastActive) > client.PingInterval() {
			err := client.WriteMessage([]byte("ping"))
			if err != nil {
				log.Printf("error pinging client: %v", err)
				m.Unregister(client)
			} else {
				m.clients[client] = time.Now()
			}
		}
	}
}

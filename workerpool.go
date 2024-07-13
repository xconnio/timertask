package workerpool

import (
	"container/heap"
	"log"
	"sync"
	"time"
)

type Client interface {
	WriteMessage(data []byte) error
	PingInterval() time.Duration
}

type ClientWrapper struct {
	Client
	NextPing time.Time
	Index    int
}

type ClientHeap []*ClientWrapper

func (h ClientHeap) Len() int {
	return len(h)
}

func (h ClientHeap) Less(i, j int) bool {
	return h[i].NextPing.Before(h[j].NextPing)
}

func (h ClientHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].Index, h[j].Index = i, j
}

// Push adds a client to the heap
func (h *ClientHeap) Push(x interface{}) {
	n := len(*h)
	client := x.(*ClientWrapper)
	client.Index = n
	*h = append(*h, client)
}

// Pop removes and returns the last client in the heap
func (h *ClientHeap) Pop() interface{} {
	old := *h
	n := len(old)
	client := old[n-1]
	client.Index = -1
	*h = old[0 : n-1]
	return client
}

type PingManager struct {
	clients   ClientHeap
	clientMap map[Client]int
	mutex     sync.Mutex
}

func NewPingManager() *PingManager {
	return &PingManager{
		clients:   make(ClientHeap, 0),
		clientMap: make(map[Client]int),
	}
}

func (m *PingManager) Add(client Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	wrappedClient := &ClientWrapper{
		Client:   client,
		NextPing: time.Now().Add(client.PingInterval()),
	}
	heap.Push(&m.clients, wrappedClient)
	m.clientMap[client] = wrappedClient.Index
}

func (m *PingManager) Remove(client Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	index, ok := m.clientMap[client]
	if ok {
		heap.Remove(&m.clients, index)
		delete(m.clientMap, client)
	}
}

func (m *PingManager) Start() {
	go func() {
		for {
			m.mutex.Lock()
			if len(m.clients) == 0 {
				m.mutex.Unlock()
				time.Sleep(1 * time.Second)
				continue
			}

			nextClient := m.clients[0]
			waitTime := time.Until(nextClient.NextPing)
			m.mutex.Unlock()

			if waitTime > 0 {
				time.Sleep(waitTime)
			}

			m.mutex.Lock()
			if time.Now().After(nextClient.NextPing) {
				err := nextClient.WriteMessage([]byte("ping"))
				if err != nil {
					log.Printf("error pinging client: %v", err)
					m.Remove(nextClient.Client)
				} else {
					nextClient.NextPing = time.Now().Add(nextClient.Client.PingInterval())
					heap.Fix(&m.clients, 0)
				}
			}
			m.mutex.Unlock()
		}
	}()
}

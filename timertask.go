package timertask

import (
	"log"
	"sync"
	"time"
)

var sleepDuration = 1 * time.Second // nolint:gochecknoglobals
var unixTimeNow = func() int64 {    // nolint:gochecknoglobals
	return time.Now().Unix()
}

type taskConfig struct {
	interval time.Duration
	callback func() error
}

type Manager struct {
	workers            map[int64]map[int64]taskConfig
	workersByIntervals map[int64]int64
	mutex              sync.Mutex
}

// NewManager creates a new Manager instance.
func NewManager() *Manager {
	return &Manager{
		workers:            make(map[int64]map[int64]taskConfig),
		workersByIntervals: make(map[int64]int64),
	}
}

// Schedule registers a new worker.
func (m *Manager) Schedule(id int64, interval time.Duration, callback func() error) {
	// Calculate the next worker time in seconds
	timeToWork := time.Now().Add(interval).Unix()

	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.workers[timeToWork] == nil {
		m.workers[timeToWork] = make(map[int64]taskConfig)
	}

	m.workers[timeToWork][id] = taskConfig{
		interval: interval,
		callback: callback,
	}
	m.workersByIntervals[id] = timeToWork
}

// Cancel unregisters a worker.
func (m *Manager) Cancel(id int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	nexTime, ok := m.workersByIntervals[id]
	if ok {
		delete(m.workers[nexTime], id)
		if len(m.workers[nexTime]) == 0 {
			delete(m.workers, nexTime)
		}
		delete(m.workersByIntervals, id)
	}
}

// Reset reschedules the worker for a specific id.
func (m *Manager) Reset(id int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	nextTime, ok := m.workersByIntervals[id]
	if ok {
		wpCfg := m.workers[nextTime][id]
		delete(m.workers[nextTime], id)
		if len(m.workers[nextTime]) == 0 {
			delete(m.workers, nextTime)
		}

		newTime := time.Now().Add(wpCfg.interval).Unix()
		if m.workers[newTime] == nil {
			m.workers[newTime] = make(map[int64]taskConfig)
		}
		m.workers[newTime][id] = wpCfg
		m.workersByIntervals[id] = newTime
	}
}

// Start begins the worker process.
func (m *Manager) Start() {
	go func() {
		for {
			nowSeconds := unixTimeNow()

			m.mutex.Lock()
			workers, ok := m.workers[nowSeconds]
			if ok {
				for id, wpCfg := range workers {
					err := wpCfg.callback()
					if err != nil {
						log.Printf("error pinging client: %v", err)
						m.Cancel(id)
						continue
					}

					// Reschedule the next worker
					timeToWork := time.Now().Add(wpCfg.interval).Unix()
					if m.workers[timeToWork] == nil {
						m.workers[timeToWork] = make(map[int64]taskConfig)
					}
					m.workers[timeToWork][id] = wpCfg
					m.workersByIntervals[id] = timeToWork

				}
				delete(m.workers, nowSeconds)
			}
			m.mutex.Unlock()

			time.Sleep(sleepDuration)
		}
	}()
}

package monitor

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type Monitor struct {
	mu    sync.RWMutex
	dirty bool

	MonitoredDevices map[string]time.Duration
	LastState        map[string]time.Time
	Notified         map[string]struct{}

	SendOffAlert func(device string, lastHeartbeat time.Time)
	SendOnAlert  func(device string, lastHeartbeat time.Time)
}

func NewMonitor(offAlert, onAlert func(device string, lastHeartbeat time.Time)) *Monitor {
	m := Monitor{
		mu:    sync.RWMutex{},
		dirty: false,

		MonitoredDevices: make(map[string]time.Duration),
		LastState:        make(map[string]time.Time),
		Notified:         make(map[string]struct{}),

		SendOffAlert: offAlert,
		SendOnAlert:  onAlert,
	}

	b, err := os.ReadFile("monitored.json")
	if err == nil {
		json.Unmarshal(b, &m.MonitoredDevices)
	}

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			m.mu.Lock()
			if !m.dirty {
				m.mu.Unlock()
				continue
			}

			data := m.MonitoredDevices
			m.dirty = false
			m.mu.Unlock()

			b, _ := json.MarshalIndent(data, "", " ")

			tmp := "monitored.json.tmp"
			os.WriteFile(tmp, b, 0644)
			os.Rename(tmp, "monitored.json")
		}
	}()

	return &m
}

func (m *Monitor) AddDevice(deviceID string, timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	fmt.Println("Now monitoring: ", deviceID, " with timeout ", timeout)
	m.MonitoredDevices[deviceID] = timeout
	m.dirty = true
}

func (m *Monitor) RemoveDevice(deviceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.MonitoredDevices, deviceID)
	m.dirty = true
}

func (m *Monitor) IsDeviceMonitored(deviceID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, ok := m.MonitoredDevices[deviceID]
	return ok
}

func (m *Monitor) Notify(deviceID string) {
	m.mu.Lock()
	fmt.Println("Device ", deviceID, " is offline, notifying...")
	m.Notified[deviceID] = struct{}{}
	lastState := m.LastState[deviceID]
	m.mu.Unlock()

	go m.SendOffAlert(deviceID, lastState)
}

func (m *Monitor) WasNotified(deviceID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, ok := m.Notified[deviceID]
	return ok
}

func (m *Monitor) LogState(deviceID string) {
	m.mu.Lock()
	lastHeartbeat := m.LastState[deviceID]
	m.LastState[deviceID] = time.Now()
	wasNotified := false
	if _, ok := m.Notified[deviceID]; ok {
		delete(m.Notified, deviceID)
		wasNotified = true
	}
	m.mu.Unlock()

	if wasNotified {
		fmt.Println("Device ", deviceID, " is back online")
		go m.SendOnAlert(deviceID, lastHeartbeat)
	}
}

func (m *Monitor) ListOff() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var off []string
	for k := range m.MonitoredDevices {
		last, ok := m.LastState[k]
		if !ok {
			continue
		}
		if time.Since(last) > m.MonitoredDevices[k] {
			off = append(off, k)
		}
	}

	return off
}

package services

import (
	"log"
	"runtime"
	"ssaiot_sim/ssaiot/adapters"
	"time"
)

type HeartbeatRecord struct {
	MemUsed uint64 `cbor:"mem_used"`
}

type MonitorRecord struct {
	Enabled bool  `cbor:"enabled"`
	Timeout int64 `cbor:"timeout,omitempty"`
}

type Health struct {
	broker       *adapters.Broker
	beatInterval time.Duration
	timeout      int64
	stop         chan struct{}
}

func NewHealth(broker *adapters.Broker, beatInterval time.Duration, timeout_s int64) *Health {
	return &Health{
		broker:       broker,
		beatInterval: beatInterval,
		timeout:      timeout_s,
	}
}

func (h *Health) EnableMonitor() error {
	err := h.broker.Publish(h.broker.Topics().Monitor(), MonitorRecord{
		Enabled: true,
		Timeout: h.timeout,
	}, false)
	return err
}

func (h *Health) DisableMonitor() error {
	err := h.broker.Publish(h.broker.Topics().Monitor(), MonitorRecord{
		Enabled: false,
	}, false)
	return err
}

func (h *Health) Start() error {
	h.stop = make(chan struct{})

	go func() {
		ticker := time.NewTicker(h.beatInterval)
		defer ticker.Stop()

		var m runtime.MemStats

		for {
			select {
			case <-ticker.C:
				runtime.ReadMemStats(&m)

				// Sys = memory obtained from the OS, the closest runtime metric to the process RSS
				err := h.broker.Publish(h.broker.Topics().Heartbeat(), HeartbeatRecord{
					MemUsed: m.Sys,
				}, false)
				if err != nil {
					log.Println(err)
				}
			case <-h.stop:
				return
			}
		}
	}()

	return nil
}

func (h *Health) Stop() {
	if h.stop != nil {
		close(h.stop)
		h.stop = nil
	}
}

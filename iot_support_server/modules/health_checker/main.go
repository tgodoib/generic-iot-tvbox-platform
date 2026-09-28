package main

import (
	"context"
	"fmt"
	"health_checker/monitor"
	"health_checker/record"
	"health_checker/storage"
	"log"
	"os"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

func main() {
	influxv2, _ := storage.NewInfluxV2Storage(
		mustEnv("INFLUXDB_URL"),
		mustEnv("INFLUXDB_TOKEN"),
		"unicamp",
		"device_status",
	)

	var client MQTT.Client

	mon := monitor.NewMonitor(
		func(device string, lastHeartbeat time.Time) {
			payload, err := monitor.GetOffAlert(device, lastHeartbeat)
			if err != nil {
				log.Printf("Error getting off alert payload: %s", err.Error())
				return
			}
			token := client.Publish("ssaiot/health/alert", 0, false, payload)
			token.Wait()
		},
		func(device string, lastHeartbeat time.Time) {
			payload, err := monitor.GetOnAlert(device, lastHeartbeat)
			if err != nil {
				log.Printf("Error getting on alert payload: %s", err.Error())
				return
			}
			token := client.Publish("ssaiot/health/alert", 0, false, payload)
			token.Wait()
		},
	)

	subscribe := func(c MQTT.Client) {
		token := c.Subscribe("ssaiot/device/+/state/monitor", 0, func(c MQTT.Client, m MQTT.Message) {
			rec, err := record.NewMonitorRecord(m)
			if err != nil {
				log.Printf("Failed to decode CBOR: %v\n", err)
				return
			}
			if rec.Enabled {
				mon.AddDevice(rec.DeviceID, time.Duration(rec.Timeout)*time.Second)
			} else {
				mon.RemoveDevice(rec.DeviceID)
			}
		})
		if token.Wait() && token.Error() != nil {
			log.Printf("Subscribe error (monitor): %v", token.Error())
		}

		token = c.Subscribe("ssaiot/device/+/state", 0, func(c MQTT.Client, m MQTT.Message) {
			rec, err := record.NewStateRecord(m)
			if err != nil {
				log.Printf("Failed to decode CBOR: %v\n", err)
				return
			}
			mon.LogState(rec.DeviceID)
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := influxv2.Write(ctx, rec); err != nil {
					log.Printf("Error writing to influx: %v", err)
				}
			}()
		})
		if token.Wait() && token.Error() != nil {
			log.Printf("Subscribe error (state): %v", token.Error())
		}

		log.Println("Subscriptions established")
	}

	opts := MQTT.NewClientOptions().AddBroker("tcp://localhost:1883")
	opts.SetClientID("health_checker")
	opts.SetOnConnectHandler(subscribe)
	opts.SetConnectionLostHandler(func(c MQTT.Client, err error) {
		log.Printf("Connection lost: %v", err)
	})

	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	fmt.Println("Connected to MQTT broker")
	fmt.Println("Listening to heartbeats...")

	go func() {
		for {
			devices := mon.ListOff()
			for _, device := range devices {
				if !mon.WasNotified(device) {
					mon.Notify(device)
				}
			}
			time.Sleep(1 * time.Second)
		}
	}()

	collectAndWrite := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cpuP, err := cpu.Percent(time.Second, false)
		if err != nil {
			log.Printf("Error reading cpu usage: %v\n", err)
			return
		}

		memV, err := mem.VirtualMemory()
		if err != nil {
			log.Printf("Error reading memory usage: %v\n", err)
			return
		}

		var temp float64
		if temps, err := host.SensorsTemperatures(); err == nil {
			for _, t := range temps {
				if strings.Contains(t.SensorKey, "cpu") || strings.Contains(t.SensorKey, "core") {
					temp = t.Temperature
					break
				}
			}
		}

		if err := influxv2.Write(ctx, record.StateRecord{
			DeviceID: "ssaiot",
			States: map[string]any{
				"memory_usage": memV.UsedPercent,
				"cpu_usage":    cpuP[0],
				"cpu_temp":     temp,
			},
			Timestamp: time.Now().Unix(),
		}); err != nil {
			log.Printf("Error writing to influx: %v", err)
		}
	}

	go func() {
		for {
			collectAndWrite()
			time.Sleep(30 * time.Second)
		}
	}()

	select {}
}

// mustEnv reads a required secret provided by systemd's EnvironmentFile (/etc/ssaiot/ssaiot.env).
func mustEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatalf("missing required environment variable %s", name)
	}
	return v
}

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sensor_data_collector/record"
	"sensor_data_collector/storage"
	"sync"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	influxv2, _ := storage.NewInfluxV2Storage(
		mustEnv("INFLUXDB_URL"),
		mustEnv("INFLUXDB_TOKEN"),
		"unicamp",
		"sensor_data",
	)

	const thresholdPerMin = 5

	var (
		mu     sync.Mutex
		counts = make(map[string]int)
	)

	var client MQTT.Client

	subscribe := func(c MQTT.Client) {
		if token := c.Subscribe("ssaiot/device/+/sensor/+/+", 0, func(c MQTT.Client, m MQTT.Message) {
			rec, err := record.NewSensorDataRecord(m)
			if err != nil {
				log.Printf("Failed to decode CBOR: %v\n", err)
				return
			}

			key := rec.DeviceID + "/" + rec.SensorID

			mu.Lock()
			counts[key]++
			count := counts[key]
			mu.Unlock()

			if count > thresholdPerMin {
				log.Println("Threshold reached, ignoring received sensor data.")

				payload, err := record.EncodeAlertRecord(&record.AlertRecord{
					Title:    "Sensor data upload threshold reached",
					Body:     fmt.Sprintf("Device %s / sensor %s tried to register data more than %d times per minute.", rec.DeviceID, rec.SensorID, thresholdPerMin),
					Type:     "sensor_data",
					DeviceID: rec.DeviceID,
				})

				if err != nil {
					log.Println("Failed to encode alert record:", err)
				} else {
					token := client.Publish("ssaiot/sensor_data/alert", 0, false, payload)
					token.Wait()
				}

				return
			}

			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				if err := influxv2.Write(ctx, rec); err != nil {
					log.Printf("Error uploading data: %v\n", err)
				}
			}()
		}); token.Wait() && token.Error() != nil {
			log.Printf("Subscribe error: %v", token.Error())
		}
		log.Println("Subscriptions established")
	}

	opts := MQTT.NewClientOptions().AddBroker("tcp://localhost:1883")
	opts.SetClientID("sensor_data_collector")
	opts.SetOnConnectHandler(subscribe)
	opts.SetConnectionLostHandler(func(c MQTT.Client, err error) {
		log.Printf("Connection lost: %v", err)
	})

	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	fmt.Println("Connected to MQTT broker")

	go func() {
		for {
			time.Sleep(1 * time.Minute)
			mu.Lock()
			counts = make(map[string]int)
			mu.Unlock()
		}
	}()

	fmt.Println("Listening to sensor data...")

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

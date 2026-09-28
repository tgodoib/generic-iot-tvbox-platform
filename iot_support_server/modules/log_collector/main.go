package main

import (
	"fmt"
	"log"
	"log_collector/record"
	"log_collector/storage"
	"os"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	file, _ := storage.NewFileStorage("./log.jsonl")
	loki, _ := storage.NewLokiStorage(mustEnv("LOKI_URL"))

	// Buffered channel serialises Loki writes so they arrive in the same
	// order the messages were received from MQTT.
	lokiCh := make(chan record.Record, 256)
	go func() {
		for rec := range lokiCh {
			if err := loki.Write(rec); err != nil {
				log.Println("Failed to write to loki:", err)
			}
		}
	}()

	var client MQTT.Client

	subscribe := func(c MQTT.Client) {
		if token := c.Subscribe("ssaiot/device/+/log", 0, func(c MQTT.Client, m MQTT.Message) {
			rec, err := record.NewRecord(m)
			if err != nil {
				log.Println("Failed to decode CBOR:", err)
				return
			}

			if record.IsOTA(&rec) && rec.Level == record.LevelError {
				payload, err := record.GetAlertRecord(&rec)
				if err != nil {
					log.Println("Failed to get ota alert record:", err)
				} else {
					token := client.Publish("ssaiot/ota_server/alert", 0, false, payload)
					token.Wait()
				}
			}

			if err := file.Write(rec); err != nil {
				log.Println("Failed to write to file:", err)
			}

			select {
			case lokiCh <- rec:
			default:
				log.Println("Loki write queue full, dropping log entry")
			}
		}); token.Wait() && token.Error() != nil {
			log.Printf("Subscribe error: %v", token.Error())
		}
		log.Println("Subscriptions established")
	}

	opts := MQTT.NewClientOptions().AddBroker("tcp://localhost:1883")
	opts.SetClientID("log_collector")
	opts.SetOnConnectHandler(subscribe)
	opts.SetConnectionLostHandler(func(c MQTT.Client, err error) {
		log.Printf("Connection lost: %v", err)
	})

	client = MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	fmt.Println("Connected to MQTT broker")
	fmt.Println("Listening to logs...")

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

package main

import (
	"fmt"
	"log"
	"notification_system/platforms"
	"notification_system/record"
	"notification_system/subscriptions"
	"os"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	registry := subscriptions.NewRegistry()

	telegram := platforms.NewTelegram(mustEnv("TELEGRAM_BOT_TOKEN"))

	registry.Register(
		telegram.NewRecipient(mustEnv("TELEGRAM_CHAT_ID")),
		[]record.AlertType{record.AlertAll},
		[]string{"*"},
	)

	subscribe := func(c MQTT.Client) {
		if token := c.Subscribe("ssaiot/+/alert", 0, func(c MQTT.Client, m MQTT.Message) {
			rec, err := record.NewAlertRecord(m)
			if err != nil {
				log.Printf("Failed to decode CBOR: %v", err)
				return
			}

			log.Printf("Received alert: %s", rec)
			go func() {
				errs := registry.Alert(rec)
				for _, err := range errs {
					log.Println(err)
				}
			}()
		}); token.Wait() && token.Error() != nil {
			log.Printf("Subscribe error: %v", token.Error())
		}
		log.Println("Subscriptions established")
	}

	opts := MQTT.NewClientOptions().AddBroker("tcp://localhost:1883")
	opts.SetClientID("notification_system")
	opts.SetOnConnectHandler(subscribe)
	opts.SetConnectionLostHandler(func(c MQTT.Client, err error) {
		log.Printf("Connection lost: %v", err)
	})

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	fmt.Println("Connected to MQTT broker")
	fmt.Println("Listening to alerts...")

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

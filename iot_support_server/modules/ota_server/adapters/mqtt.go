package adapters

import (
	"fmt"
	"ota_server/config"
	"ota_server/model"

	pahoMQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/fxamacker/cbor/v2"
)

type MQTT struct {
	client pahoMQTT.Client
}

func NewMQTTAdapter(config config.MQTTConfig) *MQTT {
	opts := pahoMQTT.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s:%s", config.Addr, config.Port))
	opts.SetClientID("ota_server")

	client := pahoMQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	fmt.Println("Connected to MQTT broker")

	return &MQTT{
		client: client,
	}
}

func (m *MQTT) SendManifestToClass(class string, manifest model.Manifest) error {
	topic := fmt.Sprintf("ssaiot/ota/class/%s/manifest", class)

	payload, err := cbor.Marshal(manifest)

	token := m.client.Publish(topic, 0, true, payload)
	token.Wait()

	return err
}

func (m *MQTT) SendManifestToID(id string, manifest model.Manifest) error {
	topic := fmt.Sprintf("ssaiot/ota/device/%s/manifest", id)

	payload, err := cbor.Marshal(manifest)

	token := m.client.Publish(topic, 0, true, payload)
	token.Wait()

	return err
}

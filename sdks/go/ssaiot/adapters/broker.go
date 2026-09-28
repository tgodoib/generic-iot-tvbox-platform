package adapters

import (
	"crypto/tls"
	"fmt"
	"ssaiot_sim/ssaiot"
	"sync"

	pahoMQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/fxamacker/cbor/v2"
)

type MessageHandler func(payload []byte)

type Broker struct {
	cfg       *ssaiot.Config
	mqtt      pahoMQTT.Client
	mu        sync.RWMutex
	onMessage map[string][]MessageHandler
}

func NewBroker(cfg *ssaiot.Config) *Broker {
	return &Broker{
		cfg:       cfg,
		onMessage: make(map[string][]MessageHandler),
	}
}

func (b *Broker) Connect() error {
	cert, err := tls.X509KeyPair([]byte(b.cfg.ClientCrt), []byte(b.cfg.ClientKey))
	if err != nil {
		return err
	}

	opts := pahoMQTT.NewClientOptions().
		AddBroker(fmt.Sprintf("ssl://%s:%d", b.cfg.ServerAddr, 8883)).
		SetClientID(b.cfg.DeviceID).
		SetTLSConfig(&tls.Config{
			Certificates: []tls.Certificate{cert},
		}).
		// Runs on every (re)connect: clean sessions drop subscriptions on the broker side
		SetOnConnectHandler(func(c pahoMQTT.Client) { b.subscribeAll() })

	b.mqtt = pahoMQTT.NewClient(opts)

	if token := b.mqtt.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (b *Broker) IsConnected() bool {
	return b.mqtt != nil && b.mqtt.IsConnected()
}

func (b *Broker) Close() {
	b.mqtt.Disconnect(250)
}

func (b *Broker) Publish(topic string, payload interface{}, retain bool) error {
	encoded, err := cbor.Marshal(payload)
	if err != nil {
		return err
	}

	token := b.mqtt.Publish(topic, 0, retain, encoded)
	token.Wait()

	return token.Error()
}

func (b *Broker) OnMessage(topic string, handler MessageHandler) {
	b.mu.Lock()
	_, exists := b.onMessage[topic]
	b.onMessage[topic] = append(b.onMessage[topic], handler)
	b.mu.Unlock()

	if !exists && b.IsConnected() {
		b.subscribe(topic)
	}
}

func (b *Broker) subscribeAll() {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for topic := range b.onMessage {
		b.subscribe(topic)
	}
}

func (b *Broker) subscribe(topic string) {
	b.mqtt.Subscribe(topic, 0, b.dispatch)
}

func (b *Broker) dispatch(_ pahoMQTT.Client, msg pahoMQTT.Message) {
	b.mu.RLock()
	handlers := b.onMessage[msg.Topic()]
	b.mu.RUnlock()

	for _, handler := range handlers {
		handler(msg.Payload())
	}
}

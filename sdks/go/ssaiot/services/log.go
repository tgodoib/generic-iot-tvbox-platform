package services

import (
	"fmt"
	"log"
	"ssaiot_sim/ssaiot/adapters"
)

type Level int

const (
	LevelInfo Level = iota
	LevelDebug
	LevelError
)

type Log struct {
	broker *adapters.Broker
}

type LogRecord struct {
	Level Level          `cbor:"level"`
	Tags  map[string]any `cbor:"tags"`
	Msg   string         `cbor:"msg"`
}

func NewLog(broker *adapters.Broker) *Log {
	return &Log{
		broker: broker,
	}
}

func (l *Log) Info(msg string, tags map[string]any) {
	log.Printf("[INFO] %s %s\n", mapString(tags), msg)
	err := l.send(LevelInfo, msg, tags)
	if err != nil {
		log.Printf("[ERROR] Could not send log to server: %s\n", err)
	}
}

func (l *Log) Debug(msg string, tags map[string]any) {
	log.Printf("[DEBUG] %s %s\n", mapString(tags), msg)
	err := l.send(LevelDebug, msg, tags)
	if err != nil {
		log.Printf("[ERROR] Could not send log to server: %s\n", err)
	}
}

func (l *Log) Error(msg string, tags map[string]any) {
	log.Printf("[ERROR] %s %s\n", mapString(tags), msg)
	err := l.send(LevelError, msg, tags)
	if err != nil {
		log.Printf("[ERROR] Could not send log to server: %s\n", err)
	}
}

func mapString(tags map[string]any) string {
	str := "{"

	for key, value := range tags {
		str += fmt.Sprintf("%s:%v ", key, value)
	}

	str += "}"

	return str
}

func (l *Log) send(level Level, msg string, tags map[string]any) error {
	if tags == nil {
		tags = map[string]any{}
	}
	err := l.broker.Publish(l.broker.Topics().Log(), LogRecord{
		Level: level,
		Tags:  tags,
		Msg:   msg,
	}, false)

	return err
}

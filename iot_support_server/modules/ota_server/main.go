package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"ota_server/adapters"
	"ota_server/config"
	"ota_server/ota"
	"syscall"
)

func main() {
	cfg := config.Load()

	// Temp
	cfg.MQTT = config.MQTTConfig{
		Addr: "localhost",
		Port: "1883",
	}
	cfg.Minio = config.MinioConfig{
		Endpoint:        mustEnv("MINIO_ENDPOINT"),
		AccessKeyID:     mustEnv("MINIO_ACCESS_KEY_ID"),
		SecretAccessKey: mustEnv("MINIO_SECRET_ACCESS_KEY"),
		BucketName:      "firmware",
	}
	cfg.HTTPListenAddr = "127.0.0.1:8080"
	cfg.ServerDomain = mustEnv("SERVER_DOMAIN")

	mqttAdapter := adapters.NewMQTTAdapter(cfg.MQTT)
	minioAdapter := adapters.NewMinioAdapter(cfg.Minio)

	otaService := ota.NewService(cfg.ServerDomain, minioAdapter, mqttAdapter)

	httpAdapter := adapters.NewHTTPAdapter(*otaService)
	httpAdapter.Listen(cfg.HTTPListenAddr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down...")
}

// mustEnv reads a required secret provided by systemd's EnvironmentFile (/etc/ssaiot/ssaiot.env).
func mustEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatalf("missing required environment variable %s", name)
	}
	return v
}

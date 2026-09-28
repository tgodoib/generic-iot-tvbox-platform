package main

import (
	"log"
	"ssaiot_sim/ssaiot"
	"ssaiot_sim/ssaiot/adapters"
	"ssaiot_sim/ssaiot/services"
)

func main() {
	cfg := ssaiot.Config{
		DeviceID:    "iot_device_sim",
		DeviceClass: "macos",
		ServerAddr:  serverAddr,
		ClientKey:   clientKey,
		ClientCrt:   clientCrt,
	}

	broker := adapters.NewBroker(&cfg)
	err := broker.Connect()
	if err != nil {
		log.Fatal(err)
	}

	web := adapters.NewWeb(&cfg)

	//health := services.NewHealth(broker, 30*time.Second, 60)
	//err = health.EnableMonitor()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//err = health.Start()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//err = health.DisableMonitor()
	//if err != nil {
	//	log.Fatal(err)
	//}

	logs := services.NewLog(broker)
	//logs.Info("Device initiated", nil)

	ota := services.NewOta(&cfg, broker, web, logs, func(filePath string) bool {
		logs.Info("Applying firmware from "+filePath, nil)
		return true
	})
	ota.Listen()

	//sensor := services.NewSensor(broker, "DHT11")
	//err = sensor.Register("temperature", rand.Intn(35), time.Now().Unix(), false)
	//if err != nil {
	//	logs.Error("Could not register measurement", nil)
	//}

	select {}
}

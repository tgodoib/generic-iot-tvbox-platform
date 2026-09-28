package config

type MQTTConfig struct {
	Addr any
	Port any
}
type MinioConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
}
type Config struct {
	MQTT           MQTTConfig
	Minio          MinioConfig
	HTTPListenAddr string
	ServerDomain   string
}

func Load() Config {
	return Config{}
}

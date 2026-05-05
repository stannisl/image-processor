package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type KafkaConfig struct {
	// Подключение
	Brokers []string
	Timeout time.Duration

	SASLMechanism string
	SASLUsername  string
	SASLPassword  string
	TLSEnabled    bool

	// Producer
	BatchSize    int
	BatchTimeout time.Duration

	// Consumer
	GroupID     string
	Concurrency int
	MaxRetries  int

	// Топики — только имена, не конфигурация
	Topics KafkaTopics
}

type KafkaTopics struct {
	Uploaded  string `env:"KAFKA_TOPIC_UPLOADED"  env-default:"image.uploaded"`
	Processed string `env:"KAFKA_TOPIC_PROCESSED" env-default:"image.processed"`
	Failed    string `env:"KAFKA_TOPIC_FAILED"    env-default:"image.failed"`
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

type ServerConfig struct {
	Host string
	Port string
}

func (s ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%s", s.Host, s.Port)
}

type Config struct {
	Postgres *PostgresConfig
	Kafka    *KafkaConfig
	Storage  *StorageConfig
	Image    *ImageConfig
	Server   *ServerConfig
}

type StorageConfig struct {
	AccessKey string
	SecretKey string
	Endpoint  string
}

type ImageConfig struct {
	WatermarkPath string
}

func (c PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Database,
	)
}
func LoadConfig() (*Config, error) {
	kafkaBrokers := getEnvOrDefault("KAFKA_BROKERS", "localhost:9092")
	if kafkaBrokers == "" {
		return nil, errors.New("config: no kafka brokers parsed")
	}
	brokers := strings.Split(kafkaBrokers, ",")

	c := &Config{
		Postgres: &PostgresConfig{
			Host:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
			Port:     getEnvOrDefault("POSTGRES_PORT", "5432"),
			User:     getEnvOrDefault("POSTGRES_USER", "postgres"),
			Password: getEnvOrDefault("POSTGRES_PASSWORD", "postgres"),
			Database: getEnvOrDefault("POSTGRES_DB", "image-processor-db"),
		},
		Kafka: &KafkaConfig{
			Brokers:       brokers,
			Timeout:       getEnvOrDefault("KAFKA_TIMEOUT", 10*time.Second),
			SASLMechanism: getEnvOrDefault("KAFKA_SASL_MECHANISM", "PLAIN"),
			SASLUsername:  getEnvOrDefault("KAFKA_SASL_USERNAME", "root"),
			SASLPassword:  getEnvOrDefault("KAFKA_SASL_PASSWORD", "root"),
			TLSEnabled:    getEnvOrDefault("KAFKA_TLS_ENABLED", false),
			BatchSize:     getEnvOrDefault("KAFKA_BATCH_SIZE", 100),
			BatchTimeout:  getEnvOrDefault("KAFKA_BATCH_TIMEOUT", 10*time.Millisecond),
			GroupID:       getEnvOrDefault("KAFKA_GROUP_ID", "image-processors"),
			Concurrency:   getEnvOrDefault("KAFKA_CONCURRENCY", 3),
			MaxRetries:    getEnvOrDefault("KAFKA_MAX_RETRIES", 3),
			Topics: KafkaTopics{
				Uploaded:  getEnvOrDefault("KAFKA_TOPIC_UPLOADED", "image.uploaded"),
				Processed: getEnvOrDefault("KAFKA_TOPIC_PROCESSED", "image.processed"),
				Failed:    getEnvOrDefault("KAFKA_TOPIC_FAILED", "image.failed"),
			},
		},
		Storage: &StorageConfig{
			AccessKey: getEnvOrDefault("STORAGE_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnvOrDefault("STORAGE_SECRET_KEY", "minioadmin"),
			Endpoint:  getEnvOrDefault("STORAGE_ENDPOINT", "localhost:9000"),
		},
		Image: &ImageConfig{
			WatermarkPath: getEnvOrDefault("IMAGE_WATERMARK_PATH", "images/watermark-wb-tech-school.png"),
		},
		Server: &ServerConfig{
			Host: getEnvOrDefault("SERVER_HOST", "localhost"),
			Port: getEnvOrDefault("SERVER_PORT", "8080"),
		},
	}

	return c, nil
}

func getEnvOrDefault[T string | int | bool | time.Duration](key string, fallback T) T {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	var result any
	var err error

	switch any(fallback).(type) {
	case string:
		result = v
	case int:
		result, err = strconv.Atoi(v)
	case bool:
		result, err = strconv.ParseBool(v)
	case time.Duration:
		result, err = time.ParseDuration(v)
	default:
		return fallback
	}

	if err != nil {
		return fallback
	}
	return result.(T)
}

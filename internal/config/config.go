package config

import (
	"fmt"
	"os"
	"strconv"
)

type APIConfig struct {
	API_PORT int
}

type AmqpConfig struct {
	AMQP_USER     string
	AMQP_PASSWORD string
	AMQP_HOST     string
	AMQP_PORT     int
}

type RedisConfig struct {
	REDIS_PORT        int
	REDIS_HOST        string
	REDIS_PASSWORD    string
	REDIS_DB          int
	REDIS_MAX_RETRIES int
	PUBSUB_CHANNEL    string
}

func (c *AmqpConfig) ToURI() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/", c.AMQP_USER, c.AMQP_PASSWORD, c.AMQP_HOST, c.AMQP_PORT)
}

var (
	API   APIConfig
	AMQP  AmqpConfig
	Redis RedisConfig
)

func Init() error {
	// api
	API.API_PORT = getEnvInt("API_PORT", 8080)

	// rabbitmq
	AMQP.AMQP_USER = getEnv("AMQP_USER", "guest")
	AMQP.AMQP_PASSWORD = getEnv("AMQP_PASSWORD", "guest")
	AMQP.AMQP_HOST = getEnv("AMQP_HOST", "localhost")
	AMQP.AMQP_PORT = getEnvInt("AMQP_PORT", 5672)

	// redis
	Redis.REDIS_HOST = getEnv("REDIS_HOST", "localhost")
	Redis.REDIS_PORT = getEnvInt("REDIS_PORT", 6379)
	Redis.REDIS_PASSWORD = getEnv("REDIS_PASSWORD", "")
	Redis.REDIS_DB = getEnvInt("REDIS_DB", 0)
	Redis.REDIS_MAX_RETRIES = getEnvInt("REDIS_MAX_RETRIES", 3)
	Redis.PUBSUB_CHANNEL = getEnv("PUBSUB_CHANNEL", "messages:new")

	return nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}

// Package config provides environment-based configuration for the chat system.
//
// Call Init() at application startup to populate the package-level variables
// (API, AMQP, Redis) from environment variables with sensible defaults.
package config

import (
	"fmt"
	"os"
	"strconv"
)

// APIConfig holds the HTTP server configuration.
type APIConfig struct {
	API_PORT int
}

// AmqpConfig holds the RabbitMQ connection configuration.
type AmqpConfig struct {
	AMQP_USER     string
	AMQP_PASSWORD string
	AMQP_HOST     string
	AMQP_PORT     int
}

// PostgresConfig holds the PostgreSQL connection configuration.
type PostgresConfig struct {
	POSTGRES_USER     string
	POSTGRES_PASSWORD string
	POSTGRES_DB       string
	POSTGRES_HOST     string
	POSTGRES_PORT     int
}

// ToURI returns the PostgreSQL connection string in URI format.
func (c *PostgresConfig) ToURI() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", c.POSTGRES_USER, c.POSTGRES_PASSWORD, c.POSTGRES_HOST, c.POSTGRES_PORT, c.POSTGRES_DB)
}

// RedisConfig holds the Redis client configuration.
type RedisConfig struct {
	REDIS_PORT        int
	REDIS_HOST        string
	REDIS_PASSWORD    string
	REDIS_DB          int
	REDIS_MAX_RETRIES int
	PUBSUB_CHANNEL    string
}

// ToURI returns the AMQP connection string in URI format.
func (c *AmqpConfig) ToURI() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/", c.AMQP_USER, c.AMQP_PASSWORD, c.AMQP_HOST, c.AMQP_PORT)
}

var (
	API      APIConfig
	AMQP     AmqpConfig
	Postgres PostgresConfig
	Redis    RedisConfig
)

// Init loads all configuration from environment variables with defaults.
func Init() error {
	API.API_PORT = getEnvInt("API_PORT", 8080)

	AMQP.AMQP_USER = getEnv("AMQP_USER", "guest")
	AMQP.AMQP_PASSWORD = getEnv("AMQP_PASSWORD", "guest")
	AMQP.AMQP_HOST = getEnv("AMQP_HOST", "localhost")
	AMQP.AMQP_PORT = getEnvInt("AMQP_PORT", 5672)

	Redis.REDIS_HOST = getEnv("REDIS_HOST", "localhost")
	Redis.REDIS_PORT = getEnvInt("REDIS_PORT", 6379)
	Redis.REDIS_PASSWORD = getEnv("REDIS_PASSWORD", "")
	Redis.REDIS_DB = getEnvInt("REDIS_DB", 0)
	Redis.REDIS_MAX_RETRIES = getEnvInt("REDIS_MAX_RETRIES", 3)
	Redis.PUBSUB_CHANNEL = getEnv("PUBSUB_CHANNEL", "messages:new")

	Postgres.POSTGRES_USER = getEnv("POSTGRES_USER", "postgres")
	Postgres.POSTGRES_PASSWORD = getEnv("POSTGRES_PASSWORD", "postgres")
	Postgres.POSTGRES_DB = getEnv("POSTGRES_DB", "chat")
	Postgres.POSTGRES_HOST = getEnv("POSTGRES_HOST", "localhost")
	Postgres.POSTGRES_PORT = getEnvInt("POSTGRES_PORT", 5432)

	return nil
}

// getEnv reads an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvInt reads an integer environment variable or returns a default value.
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

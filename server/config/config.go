package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	CAN      CANConfig
}

type ServerConfig struct {
	Port int
}

type DatabaseConfig struct {
	DSN string
}

type CANConfig struct {
	SerialPort   string
	BaudRate     int
	TCPHost      string
	TCPPort      int
	TCPTimeoutMS int
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8080,
		},
		Database: DatabaseConfig{
			DSN: getEnv("DB_DSN", "postgres://admin:123456@127.0.0.1:5433/testdb?sslmode=disable"),
		},
		CAN: CANConfig{
			SerialPort:   getEnv("CAN_SERIAL_PORT", "/dev/ttyUSB0"),
			BaudRate:     getEnvInt("CAN_BAUD_RATE", 115200),
			TCPHost:      getEnv("CAN_TCP_HOST", "127.0.0.1"),
			TCPPort:      getEnvInt("CAN_TCP_PORT", 9000),
			TCPTimeoutMS: getEnvInt("CAN_TCP_TIMEOUT_MS", 3000),
		},
	}
}

func (c *Config) ServerAddr() string {
	return fmt.Sprintf(":%d", c.Server.Port)
}

func (c CANConfig) TCPAddr() string {
	return fmt.Sprintf("%s:%d", c.TCPHost, c.TCPPort)
}

func (c CANConfig) TCPTimeout() time.Duration {
	return time.Duration(c.TCPTimeoutMS) * time.Millisecond
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port         int
	NodeID       string
	RaftGrpcPort int
	DataDir      string
	Bootstrap    bool
	DBType       string
	DBName       string
}

func Load() *Config {
	return &Config{
		Port:         getEnvInt("PORT", 8080),
		NodeID:       getEnv("NODE_ID", "node1"),
		RaftGrpcPort: getEnvInt("RAFT_GRPC_PORT", 9090),
		DataDir:      getEnv("DATA_DIR", "./data"),
		Bootstrap:    getEnvBool("BOOTSTRAP", false),
		DBType:       getEnv("DB_TYPE", "memory"),
		DBName:       getEnv("DB_NAME", "test_db"),
	}
}

func (c *Config) APIAddr() string {
	return fmt.Sprintf(":%d", c.Port)
}

func (c *Config) RaftAddr() string {
	return fmt.Sprintf(":%d", c.RaftGrpcPort)
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

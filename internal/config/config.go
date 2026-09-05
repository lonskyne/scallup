package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	NodeID       int
	Port         int
	RaftGrpcPort int
	Peers        map[int]string
	DataDir      string
	DBType       string
	DBName       string
}

func Load() *Config {
	return &Config{
		NodeID:       getEnvInt("NODE_ID", 0),
		Port:         getEnvInt("PORT", 8080),
		RaftGrpcPort: getEnvInt("RAFT_GRPC_PORT", 9090),
		Peers:        getEnvPeers("PEERS"),
		DataDir:      getEnv("DATA_DIR", "/home/lonskyne/scallup_data"),
		DBType:       getEnv("DB_TYPE", "jsonfile"),
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

func getEnvPeers(key string) map[int]string {
	value := os.Getenv(key)

	peers := make(map[int]string)

	if value == "" {
		return peers
	}

	for peer := range strings.SplitSeq(value, ",") {
		parts := strings.SplitN(peer, "=", 2)

		if len(parts) != 2 {
			continue
		}

		id, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		peers[id] = parts[1]
	}

	return peers
}

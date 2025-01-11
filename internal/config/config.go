package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	ServerAddress string
	KubeConfig    string
	LogLevel      string
	Environment   string
	K8sHost       string
}

// Load returns a Config struct populated with values from environment variables
func Load() (*Config, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: .env file not found, using environment variables\n")
	}

	// Get kubeconfig path
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		// Default to Windows path if not specified
		userHome, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %v", err)
		}
		kubeconfig = filepath.Join(userHome, ".kube", "config")
	}

	// Convert path separators for Windows
	kubeconfig = filepath.Clean(kubeconfig)

	// Verify kubeconfig file exists
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		return nil, fmt.Errorf("kubeconfig file not found at %s: %v", kubeconfig, err)
	}

	return &Config{
		ServerAddress: getEnv("SERVER_ADDRESS", ":8080"),
		KubeConfig:    kubeconfig,
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		Environment:   getEnv("ENV", "development"),
		K8sHost:       getEnv("K8S_HOST", "http://localhost:9000"),
	}, nil
}

// getEnv retrieves an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

package kubernetes

import (
	"fmt"
	"log"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Client wraps the Kubernetes clientset
type Client struct {
	*kubernetes.Clientset
	logger *log.Logger
}

// NewClient creates a new Kubernetes client
func NewClient(kubeconfigPath string) (*Client, error) {
	logger := log.New(os.Stdout, "[K8S-CLIENT] ", log.LstdFlags)

	// Get K8s host and token from environment
	k8sHost := os.Getenv("K8S_HOST")
	k8sToken := os.Getenv("K8S_TOKEN")

	if k8sHost == "" {
		return nil, fmt.Errorf("K8S_HOST environment variable is required")
	}

	if k8sToken == "" {
		return nil, fmt.Errorf("K8S_TOKEN environment variable is required")
	}

	// Create config for token-based authentication
	config := &rest.Config{
		Host:        k8sHost,
		BearerToken: k8sToken,
		// Skip TLS verification for local development
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}

	logger.Printf("Connecting to Kubernetes API at: %s", k8sHost)

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	client := &Client{
		Clientset: clientset,
		logger:    logger,
	}

	// Test the connection
	if err := client.IsHealthy(); err != nil {
		return nil, fmt.Errorf("failed health check: %v", err)
	}

	return client, nil
}

// IsHealthy checks if the connection to the Kubernetes cluster is healthy
func (c *Client) IsHealthy() error {
	version, err := c.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to Kubernetes API: %v", err)
	}
	c.logger.Printf("Successfully connected to Kubernetes %s", version.String())
	return nil
}

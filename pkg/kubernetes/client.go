package kubernetes

import (
	"fmt"
	"log"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// Client wraps the Kubernetes clientset
type Client struct {
	*kubernetes.Clientset
	logger *log.Logger
}

// NewClient creates a new Kubernetes client
func NewClient(kubeconfigPath string) (*Client, error) {
	logger := log.New(os.Stdout, "[K8S-CLIENT] ", log.LstdFlags)

	logger.Printf("Loading kubeconfig from: %s", kubeconfigPath)

	// Load the kubeconfig file
	configLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{
			ClusterInfo: clientcmdapi.Cluster{
				Server: os.Getenv("K8S_HOST"),
			},
		})

	// Get Config
	config, err := configLoader.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get client config: %v", err)
	}

	logger.Printf("Using API Server: %s", config.Host)

	// Log if we have client certificate data
	if len(config.TLSClientConfig.CertData) > 0 {
		logger.Printf("Client certificate data is present")
	}
	if len(config.TLSClientConfig.KeyData) > 0 {
		logger.Printf("Client key data is present")
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	// Test the connection
	client := &Client{
		Clientset: clientset,
		logger:    logger,
	}

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

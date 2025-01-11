// internal/api/namespace/namespace.go
package namespace

import (
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s-glance-backend/internal/api/base"
)

// NamespaceAPI handles namespace-related operations
type NamespaceAPI struct {
	*base.BaseAPI
}

// NewNamespaceAPI creates a new NamespaceAPI instance
func NewNamespaceAPI(clientset *kubernetes.Clientset, logger *log.Logger) *NamespaceAPI {
	return &NamespaceAPI{
		BaseAPI: base.NewBaseAPI(clientset, logger),
	}
}

// ListNamespaces returns all namespaces
func (api *NamespaceAPI) ListNamespaces(ctx context.Context) (*corev1.NamespaceList, error) {
	api.LogInfo(ctx, "ListNamespaces", "Fetching all namespaces")

	namespaces, err := api.GetClientset().CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		api.LogError(ctx, "ListNamespaces", err)
		return nil, api.HandleError(err, "list namespaces")
	}

	return namespaces, nil
}

// GetNamespace returns a specific namespace
func (api *NamespaceAPI) GetNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	api.LogInfo(ctx, "GetNamespace", fmt.Sprintf("Fetching namespace: %s", name))

	namespace, err := api.GetClientset().CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		api.LogError(ctx, "GetNamespace", err)
		return nil, api.HandleError(err, "get namespace")
	}

	return namespace, nil
}

// GetNamespaceMetrics returns resource usage for a namespace
func (api *NamespaceAPI) GetNamespaceMetrics(ctx context.Context, name string) (*base.APIResponse, error) {
	api.LogInfo(ctx, "GetNamespaceMetrics", fmt.Sprintf("Fetching metrics for namespace: %s", name))

	// Get pods in namespace
	pods, err := api.GetClientset().CoreV1().Pods(name).List(ctx, metav1.ListOptions{})
	if err != nil {
		api.LogError(ctx, "GetNamespaceMetrics", err)
		return nil, api.HandleError(err, "get namespace metrics")
	}

	// Calculate resource usage
	metrics := map[string]interface{}{
		"podCount": len(pods.Items),
		"status": map[string]int{
			"running":   0,
			"pending":   0,
			"failed":    0,
			"succeeded": 0,
		},
	}

	// Count pods by status
	for _, pod := range pods.Items {
		metrics["status"].(map[string]int)[string(pod.Status.Phase)]++
	}

	response := base.NewSuccessResponse(metrics)
	return &response, nil
}

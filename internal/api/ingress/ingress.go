package ingress

import (
	"context"
	"fmt"
	"log"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s-glance-backend/internal/api/base"
)

// IngressAPI handles ingress-related operations
type IngressAPI struct {
	*base.BaseAPI
}

// NewIngressAPI creates a new IngressAPI instance
func NewIngressAPI(clientset *kubernetes.Clientset, logger *log.Logger) *IngressAPI {
	return &IngressAPI{
		BaseAPI: base.NewBaseAPI(clientset, logger),
	}
}

// ListIngresses returns all ingresses in a namespace
func (api *IngressAPI) ListIngresses(ctx context.Context, namespace string) (*networkingv1.IngressList, error) {
	api.LogInfo(ctx, "ListIngresses", fmt.Sprintf("Fetching ingresses in namespace: %s", namespace))

	ingresses, err := api.GetClientset().NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		api.LogError(ctx, "ListIngresses", err)
		return nil, api.HandleError(err, "list ingresses")
	}

	return ingresses, nil
}

// GetIngress returns a specific ingress
func (api *IngressAPI) GetIngress(ctx context.Context, namespace, name string) (*networkingv1.Ingress, error) {
	api.LogInfo(ctx, "GetIngress", fmt.Sprintf("Fetching ingress %s in namespace %s", name, namespace))

	ingress, err := api.GetClientset().NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		api.LogError(ctx, "GetIngress", err)
		return nil, api.HandleError(err, "get ingress")
	}

	return ingress, nil
}

// CreateIngress creates a new ingress
func (api *IngressAPI) CreateIngress(ctx context.Context, namespace string, ingress *networkingv1.Ingress) (*networkingv1.Ingress, error) {
	api.LogInfo(ctx, "CreateIngress", fmt.Sprintf("Creating ingress %s in namespace %s", ingress.Name, namespace))

	result, err := api.GetClientset().NetworkingV1().Ingresses(namespace).Create(ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		api.LogError(ctx, "CreateIngress", err)
		return nil, api.HandleError(err, "create ingress")
	}

	return result, nil
}

// UpdateIngress updates an existing ingress
func (api *IngressAPI) UpdateIngress(ctx context.Context, namespace string, ingress *networkingv1.Ingress) (*networkingv1.Ingress, error) {
	api.LogInfo(ctx, "UpdateIngress", fmt.Sprintf("Updating ingress %s in namespace %s", ingress.Name, namespace))

	result, err := api.GetClientset().NetworkingV1().Ingresses(namespace).Update(ctx, ingress, metav1.UpdateOptions{})
	if err != nil {
		api.LogError(ctx, "UpdateIngress", err)
		return nil, api.HandleError(err, "update ingress")
	}

	return result, nil
}

// DeleteIngress deletes a specific ingress
func (api *IngressAPI) DeleteIngress(ctx context.Context, namespace, name string) error {
	api.LogInfo(ctx, "DeleteIngress", fmt.Sprintf("Deleting ingress %s in namespace %s", name, namespace))

	err := api.GetClientset().NetworkingV1().Ingresses(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		api.LogError(ctx, "DeleteIngress", err)
		return api.HandleError(err, "delete ingress")
	}

	return nil
}

// GetIngressStatus returns detailed status of an ingress
func (api *IngressAPI) GetIngressStatus(ctx context.Context, namespace, name string) (*base.APIResponse, error) {
	api.LogInfo(ctx, "GetIngressStatus", fmt.Sprintf("Fetching status for ingress %s in namespace %s", name, namespace))

	ingress, err := api.GetIngress(ctx, namespace, name)
	if err != nil {
		return nil, err
	}

	// Build detailed status response
	status := map[string]interface{}{
		"loadBalancer": getLoadBalancerStatus(ingress.Status.LoadBalancer),
		"rules":        getIngressRules(ingress.Spec.Rules),
		"tls":          getTLSStatus(ingress.Spec.TLS),
		"class":        ingress.Spec.IngressClassName,
		"annotations":  ingress.Annotations,
	}

	response := base.NewSuccessResponse(status)
	return &response, nil
}

// Helper functions

func getLoadBalancerStatus(status networkingv1.IngressLoadBalancerStatus) []map[string]interface{} {
	var result []map[string]interface{}
	for _, ingress := range status.Ingress {
		result = append(result, map[string]interface{}{
			"ip":       ingress.IP,
			"hostname": ingress.Hostname,
		})
	}
	return result
}

func getIngressRules(rules []networkingv1.IngressRule) []map[string]interface{} {
	var result []map[string]interface{}
	for _, rule := range rules {
		paths := make([]map[string]interface{}, 0)
		if rule.HTTP != nil {
			for _, path := range rule.HTTP.Paths {
				paths = append(paths, map[string]interface{}{
					"path":     path.Path,
					"pathType": path.PathType,
					"backend": map[string]interface{}{
						"service": map[string]interface{}{
							"name": path.Backend.Service.Name,
							"port": path.Backend.Service.Port,
						},
					},
				})
			}
		}

		result = append(result, map[string]interface{}{
			"host":  rule.Host,
			"paths": paths,
		})
	}
	return result
}

func getTLSStatus(tls []networkingv1.IngressTLS) []map[string]interface{} {
	var result []map[string]interface{}
	for _, t := range tls {
		result = append(result, map[string]interface{}{
			"hosts":      t.Hosts,
			"secretName": t.SecretName,
		})
	}
	return result
}

package service

import (
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s-glance-backend/internal/api/base"
)

// ServiceAPI handles service-related operations
type ServiceAPI struct {
	*base.BaseAPI
}

// NewServiceAPI creates a new ServiceAPI instance
func NewServiceAPI(clientset *kubernetes.Clientset, logger *log.Logger) *ServiceAPI {
	return &ServiceAPI{
		BaseAPI: base.NewBaseAPI(clientset, logger),
	}
}

// ListServices returns all services in a namespace
func (api *ServiceAPI) ListServices(ctx context.Context, namespace string) (*corev1.ServiceList, error) {
	api.LogInfo(ctx, "ListServices", fmt.Sprintf("Fetching services in namespace: %s", namespace))

	services, err := api.GetClientset().CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		api.LogError(ctx, "ListServices", err)
		return nil, api.HandleError(err, "list services")
	}

	return services, nil
}

// GetService returns a specific service
func (api *ServiceAPI) GetService(ctx context.Context, namespace, name string) (*corev1.Service, error) {
	api.LogInfo(ctx, "GetService", fmt.Sprintf("Fetching service %s in namespace %s", name, namespace))

	service, err := api.GetClientset().CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		api.LogError(ctx, "GetService", err)
		return nil, api.HandleError(err, "get service")
	}

	return service, nil
}

// CreateService creates a new service
func (api *ServiceAPI) CreateService(ctx context.Context, namespace string, service *corev1.Service) (*corev1.Service, error) {
	api.LogInfo(ctx, "CreateService", fmt.Sprintf("Creating service %s in namespace %s", service.Name, namespace))

	result, err := api.GetClientset().CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		api.LogError(ctx, "CreateService", err)
		return nil, api.HandleError(err, "create service")
	}

	return result, nil
}

// UpdateService updates an existing service
func (api *ServiceAPI) UpdateService(ctx context.Context, namespace string, service *corev1.Service) (*corev1.Service, error) {
	api.LogInfo(ctx, "UpdateService", fmt.Sprintf("Updating service %s in namespace %s", service.Name, namespace))

	result, err := api.GetClientset().CoreV1().Services(namespace).Update(ctx, service, metav1.UpdateOptions{})
	if err != nil {
		api.LogError(ctx, "UpdateService", err)
		return nil, api.HandleError(err, "update service")
	}

	return result, nil
}

// DeleteService deletes a specific service
func (api *ServiceAPI) DeleteService(ctx context.Context, namespace, name string) error {
	api.LogInfo(ctx, "DeleteService", fmt.Sprintf("Deleting service %s in namespace %s", name, namespace))

	err := api.GetClientset().CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		api.LogError(ctx, "DeleteService", err)
		return api.HandleError(err, "delete service")
	}

	return nil
}

// GetServiceStatus returns the status and endpoints of a service
func (api *ServiceAPI) GetServiceStatus(ctx context.Context, namespace, name string) (*base.APIResponse, error) {
	api.LogInfo(ctx, "GetServiceStatus", fmt.Sprintf("Fetching status for service %s in namespace %s", name, namespace))

	// Get service details
	service, err := api.GetService(ctx, namespace, name)
	if err != nil {
		return nil, err
	}

	// Get endpoints for the service
	endpoints, err := api.GetClientset().CoreV1().Endpoints(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		api.LogError(ctx, "GetServiceStatus", err)
		return nil, api.HandleError(err, "get service endpoints")
	}

	// Collect LoadBalancer status if applicable
	var lbStatus []map[string]interface{}
	if service.Spec.Type == corev1.ServiceTypeLoadBalancer {
		for _, ingress := range service.Status.LoadBalancer.Ingress {
			lbStatus = append(lbStatus, map[string]interface{}{
				"ip":       ingress.IP,
				"hostname": ingress.Hostname,
			})
		}
	}

	// Build detailed status response
	status := map[string]interface{}{
		"type":            service.Spec.Type,
		"clusterIP":       service.Spec.ClusterIP,
		"externalIPs":     service.Spec.ExternalIPs,
		"loadBalancer":    lbStatus,
		"ports":           getServicePorts(service.Spec.Ports),
		"endpoints":       getEndpointAddresses(endpoints),
		"selector":        service.Spec.Selector,
		"sessionAffinity": string(service.Spec.SessionAffinity),
	}

	response := base.NewSuccessResponse(status)
	return &response, nil
}

// Helper functions

func getServicePorts(ports []corev1.ServicePort) []map[string]interface{} {
	var result []map[string]interface{}
	for _, port := range ports {
		result = append(result, map[string]interface{}{
			"name":       port.Name,
			"protocol":   string(port.Protocol),
			"port":       port.Port,
			"targetPort": port.TargetPort.String(),
			"nodePort":   port.NodePort,
		})
	}
	return result
}

func getEndpointAddresses(endpoints *corev1.Endpoints) []map[string]interface{} {
	var result []map[string]interface{}
	for _, subset := range endpoints.Subsets {
		for _, addr := range subset.Addresses {
			result = append(result, map[string]interface{}{
				"ip":        addr.IP,
				"hostname":  addr.Hostname,
				"nodeName":  addr.NodeName,
				"targetRef": addr.TargetRef,
			})
		}
	}
	return result
}

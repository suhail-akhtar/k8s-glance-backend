package configmap

import (
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s-glance-backend/internal/api/base"
)

// ConfigMapAPI handles ConfigMap-related operations
type ConfigMapAPI struct {
	*base.BaseAPI
}

// NewConfigMapAPI creates a new ConfigMapAPI instance
func NewConfigMapAPI(clientset *kubernetes.Clientset, logger *log.Logger) *ConfigMapAPI {
	return &ConfigMapAPI{
		BaseAPI: base.NewBaseAPI(clientset, logger),
	}
}

// ListConfigMaps returns all ConfigMaps in a namespace
func (api *ConfigMapAPI) ListConfigMaps(ctx context.Context, namespace string) (*corev1.ConfigMapList, error) {
	api.LogInfo(ctx, "ListConfigMaps", fmt.Sprintf("Fetching ConfigMaps in namespace: %s", namespace))

	configMaps, err := api.GetClientset().CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		api.LogError(ctx, "ListConfigMaps", err)
		return nil, api.HandleError(err, "list configmaps")
	}

	return configMaps, nil
}

// GetConfigMap returns a specific ConfigMap
func (api *ConfigMapAPI) GetConfigMap(ctx context.Context, namespace, name string) (*corev1.ConfigMap, error) {
	api.LogInfo(ctx, "GetConfigMap", fmt.Sprintf("Fetching ConfigMap %s in namespace %s", name, namespace))

	configMap, err := api.GetClientset().CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		api.LogError(ctx, "GetConfigMap", err)
		return nil, api.HandleError(err, "get configmap")
	}

	return configMap, nil
}

// CreateConfigMap creates a new ConfigMap
func (api *ConfigMapAPI) CreateConfigMap(ctx context.Context, namespace string, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	api.LogInfo(ctx, "CreateConfigMap", fmt.Sprintf("Creating ConfigMap %s in namespace %s", configMap.Name, namespace))

	result, err := api.GetClientset().CoreV1().ConfigMaps(namespace).Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		api.LogError(ctx, "CreateConfigMap", err)
		return nil, api.HandleError(err, "create configmap")
	}

	return result, nil
}

// UpdateConfigMap updates an existing ConfigMap
func (api *ConfigMapAPI) UpdateConfigMap(ctx context.Context, namespace string, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	api.LogInfo(ctx, "UpdateConfigMap", fmt.Sprintf("Updating ConfigMap %s in namespace %s", configMap.Name, namespace))

	result, err := api.GetClientset().CoreV1().ConfigMaps(namespace).Update(ctx, configMap, metav1.UpdateOptions{})
	if err != nil {
		api.LogError(ctx, "UpdateConfigMap", err)
		return nil, api.HandleError(err, "update configmap")
	}

	return result, nil
}

// DeleteConfigMap deletes a specific ConfigMap
func (api *ConfigMapAPI) DeleteConfigMap(ctx context.Context, namespace, name string) error {
	api.LogInfo(ctx, "DeleteConfigMap", fmt.Sprintf("Deleting ConfigMap %s in namespace %s", name, namespace))

	err := api.GetClientset().CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		api.LogError(ctx, "DeleteConfigMap", err)
		return api.HandleError(err, "delete configmap")
	}

	return nil
}

// GetConfigMapUsage returns information about which Pods are using this ConfigMap
func (api *ConfigMapAPI) GetConfigMapUsage(ctx context.Context, namespace, name string) (*base.APIResponse, error) {
	api.LogInfo(ctx, "GetConfigMapUsage", fmt.Sprintf("Checking usage of ConfigMap %s in namespace %s", name, namespace))

	// Get pods in the namespace
	pods, err := api.GetClientset().CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		api.LogError(ctx, "GetConfigMapUsage", err)
		return nil, api.HandleError(err, "list pods for configmap usage")
	}

	// Track pods using this ConfigMap
	var usingPods []map[string]interface{}

	for _, pod := range pods.Items {
		isUsed := false
		usageDetails := make(map[string][]string)

		// Check volume mounts
		for _, volume := range pod.Spec.Volumes {
			if volume.ConfigMap != nil && volume.ConfigMap.Name == name {
				isUsed = true
				usageDetails["volumeMounts"] = append(usageDetails["volumeMounts"], volume.Name)
			}
		}

		// Check environment variables
		for _, container := range pod.Spec.Containers {
			for _, env := range container.EnvFrom {
				if env.ConfigMapRef != nil && env.ConfigMapRef.Name == name {
					isUsed = true
					usageDetails["envFrom"] = append(usageDetails["envFrom"], container.Name)
				}
			}

			for _, env := range container.Env {
				if env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil &&
					env.ValueFrom.ConfigMapKeyRef.Name == name {
					isUsed = true
					usageDetails["envVars"] = append(usageDetails["envVars"],
						fmt.Sprintf("%s:%s", container.Name, env.Name))
				}
			}
		}

		if isUsed {
			usingPods = append(usingPods, map[string]interface{}{
				"name":   pod.Name,
				"status": pod.Status.Phase,
				"usage":  usageDetails,
			})
		}
	}

	response := base.NewSuccessResponse(map[string]interface{}{
		"podsUsingConfigMap": usingPods,
		"totalPods":          len(usingPods),
	})
	return &response, nil
}

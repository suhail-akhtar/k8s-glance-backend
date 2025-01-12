package secret

import (
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s-glance-backend/internal/api/base"
)

// SecretAPI handles Secret-related operations
type SecretAPI struct {
	*base.BaseAPI
}

// NewSecretAPI creates a new SecretAPI instance
func NewSecretAPI(clientset *kubernetes.Clientset, logger *log.Logger) *SecretAPI {
	return &SecretAPI{
		BaseAPI: base.NewBaseAPI(clientset, logger),
	}
}

// ListSecrets returns all Secrets in a namespace (without their values)
func (api *SecretAPI) ListSecrets(ctx context.Context, namespace string) (*corev1.SecretList, error) {
	api.LogInfo(ctx, "ListSecrets", fmt.Sprintf("Fetching Secrets in namespace: %s", namespace))

	secrets, err := api.GetClientset().CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		api.LogError(ctx, "ListSecrets", err)
		return nil, api.HandleError(err, "list secrets")
	}

	// Remove sensitive data before returning
	for i := range secrets.Items {
		secrets.Items[i].Data = nil
		secrets.Items[i].StringData = nil
	}

	return secrets, nil
}

// GetSecret returns metadata about a specific Secret (without values)
func (api *SecretAPI) GetSecret(ctx context.Context, namespace, name string) (*corev1.Secret, error) {
	api.LogInfo(ctx, "GetSecret", fmt.Sprintf("Fetching Secret %s in namespace %s", name, namespace))

	secret, err := api.GetClientset().CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		api.LogError(ctx, "GetSecret", err)
		return nil, api.HandleError(err, "get secret")
	}

	// Remove sensitive data before returning
	secret.Data = nil
	secret.StringData = nil

	return secret, nil
}

// GetSecretKeys returns only the keys (not values) of a Secret
func (api *SecretAPI) GetSecretKeys(ctx context.Context, namespace, name string) (*base.APIResponse, error) {
	api.LogInfo(ctx, "GetSecretKeys", fmt.Sprintf("Fetching keys for Secret %s in namespace %s", name, namespace))

	secret, err := api.GetClientset().CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		api.LogError(ctx, "GetSecretKeys", err)
		return nil, api.HandleError(err, "get secret keys")
	}

	// Extract only the keys
	keys := make([]string, 0, len(secret.Data))
	for k := range secret.Data {
		keys = append(keys, k)
	}

	response := base.NewSuccessResponse(map[string]interface{}{
		"keys": keys,
		"type": string(secret.Type),
	})
	return &response, nil
}

// CreateSecret creates a new Secret
func (api *SecretAPI) CreateSecret(ctx context.Context, namespace string, secret *corev1.Secret) (*corev1.Secret, error) {
	api.LogInfo(ctx, "CreateSecret", fmt.Sprintf("Creating Secret %s in namespace %s", secret.Name, namespace))

	result, err := api.GetClientset().CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		api.LogError(ctx, "CreateSecret", err)
		return nil, api.HandleError(err, "create secret")
	}

	// Remove sensitive data before returning
	result.Data = nil
	result.StringData = nil

	return result, nil
}

// UpdateSecret updates an existing Secret
func (api *SecretAPI) UpdateSecret(ctx context.Context, namespace string, secret *corev1.Secret) (*corev1.Secret, error) {
	api.LogInfo(ctx, "UpdateSecret", fmt.Sprintf("Updating Secret %s in namespace %s", secret.Name, namespace))

	result, err := api.GetClientset().CoreV1().Secrets(namespace).Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		api.LogError(ctx, "UpdateSecret", err)
		return nil, api.HandleError(err, "update secret")
	}

	// Remove sensitive data before returning
	result.Data = nil
	result.StringData = nil

	return result, nil
}

// DeleteSecret deletes a specific Secret
func (api *SecretAPI) DeleteSecret(ctx context.Context, namespace, name string) error {
	api.LogInfo(ctx, "DeleteSecret", fmt.Sprintf("Deleting Secret %s in namespace %s", name, namespace))

	err := api.GetClientset().CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		api.LogError(ctx, "DeleteSecret", err)
		return api.HandleError(err, "delete secret")
	}

	return nil
}

// GetSecretUsage returns information about which Pods are using this Secret
func (api *SecretAPI) GetSecretUsage(ctx context.Context, namespace, name string) (*base.APIResponse, error) {
	api.LogInfo(ctx, "GetSecretUsage", fmt.Sprintf("Checking usage of Secret %s in namespace %s", name, namespace))

	pods, err := api.GetClientset().CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		api.LogError(ctx, "GetSecretUsage", err)
		return nil, api.HandleError(err, "list pods for secret usage")
	}

	var usingPods []map[string]interface{}

	for _, pod := range pods.Items {
		isUsed := false
		usageDetails := make(map[string][]string)

		// Check volumes
		for _, volume := range pod.Spec.Volumes {
			if volume.Secret != nil && volume.Secret.SecretName == name {
				isUsed = true
				usageDetails["volumeMounts"] = append(usageDetails["volumeMounts"], volume.Name)
			}
		}

		// Check environment variables
		for _, container := range pod.Spec.Containers {
			for _, env := range container.EnvFrom {
				if env.SecretRef != nil && env.SecretRef.Name == name {
					isUsed = true
					usageDetails["envFrom"] = append(usageDetails["envFrom"], container.Name)
				}
			}

			for _, env := range container.Env {
				if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil &&
					env.ValueFrom.SecretKeyRef.Name == name {
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
		"podsUsingSecret": usingPods,
		"totalPods":       len(usingPods),
	})
	return &response, nil
}

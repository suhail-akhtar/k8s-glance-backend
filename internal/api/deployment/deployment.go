package deployment

import (
	"context"
	"fmt"
	"log"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s-glance-backend/internal/api/base"
)

// DeploymentAPI handles deployment-related operations
type DeploymentAPI struct {
	*base.BaseAPI
}

// NewDeploymentAPI creates a new DeploymentAPI instance
func NewDeploymentAPI(clientset *kubernetes.Clientset, logger *log.Logger) *DeploymentAPI {
	return &DeploymentAPI{
		BaseAPI: base.NewBaseAPI(clientset, logger),
	}
}

// ListDeployments returns all deployments in a namespace
func (api *DeploymentAPI) ListDeployments(ctx context.Context, namespace string) (*appsv1.DeploymentList, error) {
	api.LogInfo(ctx, "ListDeployments", fmt.Sprintf("Fetching deployments in namespace: %s", namespace))

	deployments, err := api.GetClientset().AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		api.LogError(ctx, "ListDeployments", err)
		return nil, api.HandleError(err, "list deployments")
	}

	return deployments, nil
}

// GetDeployment returns a specific deployment
func (api *DeploymentAPI) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	api.LogInfo(ctx, "GetDeployment", fmt.Sprintf("Fetching deployment %s in namespace %s", name, namespace))

	deployment, err := api.GetClientset().AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		api.LogError(ctx, "GetDeployment", err)
		return nil, api.HandleError(err, "get deployment")
	}

	return deployment, nil
}

// GetDeploymentStatus returns detailed status of a deployment
func (api *DeploymentAPI) GetDeploymentStatus(ctx context.Context, namespace, name string) (*base.APIResponse, error) {
	api.LogInfo(ctx, "GetDeploymentStatus", fmt.Sprintf("Fetching status for deployment %s in namespace %s", name, namespace))

	deployment, err := api.GetDeployment(ctx, namespace, name)
	if err != nil {
		return nil, err
	}

	status := map[string]interface{}{
		"replicas": map[string]int32{
			"desired":   *deployment.Spec.Replicas,
			"current":   deployment.Status.Replicas,
			"updated":   deployment.Status.UpdatedReplicas,
			"ready":     deployment.Status.ReadyReplicas,
			"available": deployment.Status.AvailableReplicas,
		},
		"conditions": getDeploymentConditions(deployment.Status.Conditions),
		"strategy":   deployment.Spec.Strategy.Type,
		"age":        deployment.CreationTimestamp.Time,
	}

	response := base.NewSuccessResponse(status)
	return &response, nil
}

// DeleteDeployment deletes a specific deployment
func (api *DeploymentAPI) DeleteDeployment(ctx context.Context, namespace, name string) error {
	api.LogInfo(ctx, "DeleteDeployment", fmt.Sprintf("Deleting deployment %s in namespace %s", name, namespace))

	err := api.GetClientset().AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		api.LogError(ctx, "DeleteDeployment", err)
		return api.HandleError(err, "delete deployment")
	}

	return nil
}

// ScaleDeployment scales a deployment to the specified number of replicas
func (api *DeploymentAPI) ScaleDeployment(ctx context.Context, namespace, name string, replicas int32) error {
	api.LogInfo(ctx, "ScaleDeployment", fmt.Sprintf("Scaling deployment %s in namespace %s to %d replicas", name, namespace, replicas))

	deployment, err := api.GetDeployment(ctx, namespace, name)
	if err != nil {
		return err
	}

	deployment.Spec.Replicas = &replicas

	_, err = api.GetClientset().AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		api.LogError(ctx, "ScaleDeployment", err)
		return api.HandleError(err, "scale deployment")
	}

	return nil
}

// CreateDeployment creates a new deployment
func (api *DeploymentAPI) CreateDeployment(ctx context.Context, namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	api.LogInfo(ctx, "CreateDeployment", fmt.Sprintf("Creating deployment %s in namespace %s", deployment.Name, namespace))

	result, err := api.GetClientset().AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		api.LogError(ctx, "CreateDeployment", err)
		return nil, api.HandleError(err, "create deployment")
	}

	return result, nil
}

// UpdateDeployment updates an existing deployment
func (api *DeploymentAPI) UpdateDeployment(ctx context.Context, namespace string, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	api.LogInfo(ctx, "UpdateDeployment", fmt.Sprintf("Updating deployment %s in namespace %s", deployment.Name, namespace))

	result, err := api.GetClientset().AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		api.LogError(ctx, "UpdateDeployment", err)
		return nil, api.HandleError(err, "update deployment")
	}

	return result, nil
}

// Helper functions

func getDeploymentConditions(conditions []appsv1.DeploymentCondition) []map[string]interface{} {
	var result []map[string]interface{}
	for _, condition := range conditions {
		result = append(result, map[string]interface{}{
			"type":               condition.Type,
			"status":             condition.Status,
			"lastUpdateTime":     condition.LastUpdateTime,
			"lastTransitionTime": condition.LastTransitionTime,
			"reason":             condition.Reason,
			"message":            condition.Message,
		})
	}
	return result
}

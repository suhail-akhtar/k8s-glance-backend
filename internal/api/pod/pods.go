package pod

import (
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"k8s-glance-backend/internal/api/base"
)

// PodAPI handles pod-related operations
type PodAPI struct {
	*base.BaseAPI
}

// NewPodAPI creates a new PodAPI instance
func NewPodAPI(clientset *kubernetes.Clientset, logger *log.Logger) *PodAPI {
	return &PodAPI{
		BaseAPI: base.NewBaseAPI(clientset, logger),
	}
}

// ListPods returns all pods in a namespace
func (api *PodAPI) ListPods(ctx context.Context, namespace string) (*corev1.PodList, error) {
	api.LogInfo(ctx, "ListPods", fmt.Sprintf("Fetching pods in namespace: %s", namespace))

	pods, err := api.GetClientset().CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		api.LogError(ctx, "ListPods", err)
		return nil, api.HandleError(err, "list pods")
	}

	return pods, nil
}

// GetPod returns a specific pod
func (api *PodAPI) GetPod(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	api.LogInfo(ctx, "GetPod", fmt.Sprintf("Fetching pod %s in namespace %s", name, namespace))

	pod, err := api.GetClientset().CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		api.LogError(ctx, "GetPod", err)
		return nil, api.HandleError(err, "get pod")
	}

	return pod, nil
}

// GetPodMetrics returns resource usage for a pod
func (api *PodAPI) GetPodMetrics(ctx context.Context, namespace, name string) (*base.APIResponse, error) {
	api.LogInfo(ctx, "GetPodMetrics", fmt.Sprintf("Fetching metrics for pod %s in namespace %s", name, namespace))

	pod, err := api.GetPod(ctx, namespace, name)
	if err != nil {
		return nil, err
	}

	// Calculate container statuses
	containerStatuses := make(map[string]interface{})
	for _, container := range pod.Status.ContainerStatuses {
		containerStatuses[container.Name] = map[string]interface{}{
			"ready":        container.Ready,
			"restartCount": container.RestartCount,
			"state":        getContainerState(container.State),
		}
	}

	// Aggregate pod metrics
	metrics := map[string]interface{}{
		"phase":            pod.Status.Phase,
		"hostIP":           pod.Status.HostIP,
		"podIP":            pod.Status.PodIP,
		"startTime":        pod.Status.StartTime,
		"containers":       containerStatuses,
		"conditions":       getPodConditions(pod.Status.Conditions),
		"resourceRequests": getResourceRequests(pod.Spec.Containers),
	}

	response := base.NewSuccessResponse(metrics)
	return &response, nil
}

// DeletePod deletes a specific pod
func (api *PodAPI) DeletePod(ctx context.Context, namespace, name string) error {
	api.LogInfo(ctx, "DeletePod", fmt.Sprintf("Deleting pod %s in namespace %s", name, namespace))

	err := api.GetClientset().CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		api.LogError(ctx, "DeletePod", err)
		return api.HandleError(err, "delete pod")
	}

	return nil
}

// Helper functions

func getContainerState(state corev1.ContainerState) string {
	if state.Running != nil {
		return "Running"
	}
	if state.Waiting != nil {
		return fmt.Sprintf("Waiting (%s)", state.Waiting.Reason)
	}
	if state.Terminated != nil {
		return fmt.Sprintf("Terminated (%s)", state.Terminated.Reason)
	}
	return "Unknown"
}

func getPodConditions(conditions []corev1.PodCondition) []map[string]interface{} {
	var result []map[string]interface{}
	for _, condition := range conditions {
		result = append(result, map[string]interface{}{
			"type":    condition.Type,
			"status":  condition.Status,
			"reason":  condition.Reason,
			"message": condition.Message,
		})
	}
	return result
}

func getResourceRequests(containers []corev1.Container) []map[string]interface{} {
	var resources []map[string]interface{}
	for _, container := range containers {
		resources = append(resources, map[string]interface{}{
			"name":   container.Name,
			"cpu":    container.Resources.Requests.Cpu().String(),
			"memory": container.Resources.Requests.Memory().String(),
		})
	}
	return resources
}

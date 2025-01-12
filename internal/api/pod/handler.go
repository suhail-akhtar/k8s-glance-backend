package pod

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type Handler struct {
	api *PodAPI
}

func NewHandler(clientset *kubernetes.Clientset, logger *log.Logger) *Handler {
	if logger == nil {
		logger = log.New(gin.DefaultWriter, "[POD-API] ", log.LstdFlags)
	}

	return &Handler{
		api: NewPodAPI(clientset, logger),
	}
}

// ListPods handles GET /api/v1/namespaces/:namespace/pods
func (h *Handler) ListPods(c *gin.Context) {
	namespace := c.Param("namespace")
	pods, err := h.api.ListPods(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Convert to a simpler response format
	var response []map[string]interface{}
	for _, pod := range pods.Items {
		response = append(response, map[string]interface{}{
			"name":         pod.Name,
			"namespace":    pod.Namespace,
			"status":       pod.Status.Phase,
			"podIP":        pod.Status.PodIP,
			"hostIP":       pod.Status.HostIP,
			"creationTime": pod.CreationTimestamp,
			"labels":       pod.Labels,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetPod handles GET /api/v1/namespaces/:namespace/pods/:name
func (h *Handler) GetPod(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	pod, err := h.api.GetPod(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"name":         pod.Name,
			"namespace":    pod.Namespace,
			"status":       pod.Status.Phase,
			"podIP":        pod.Status.PodIP,
			"hostIP":       pod.Status.HostIP,
			"creationTime": pod.CreationTimestamp,
			"labels":       pod.Labels,
			"annotations":  pod.Annotations,
			"nodeName":     pod.Spec.NodeName,
			"containers":   getContainerInfo(pod.Spec.Containers),
		},
	})
}

// GetPodMetrics handles GET /api/v1/namespaces/:namespace/pods/:name/metrics
func (h *Handler) GetPodMetrics(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	metrics, err := h.api.GetPodMetrics(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// DeletePod handles DELETE /api/v1/namespaces/:namespace/pods/:name
func (h *Handler) DeletePod(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	err := h.api.DeletePod(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Pod deleted successfully",
	})
}

// Helper functions

func getContainerInfo(containers []corev1.Container) []map[string]interface{} {
	var containerInfo []map[string]interface{}
	for _, container := range containers {
		containerInfo = append(containerInfo, map[string]interface{}{
			"name":    container.Name,
			"image":   container.Image,
			"ports":   container.Ports,
			"env":     container.Env,
			"command": container.Command,
			"args":    container.Args,
		})
	}
	return containerInfo
}

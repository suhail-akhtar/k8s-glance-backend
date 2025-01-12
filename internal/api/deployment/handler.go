package deployment

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Handler struct {
	api *DeploymentAPI
}

// CreateDeployment handles POST /api/v1/deployments/namespaces/:namespace
func (h *Handler) CreateDeployment(c *gin.Context) {
	var deploymentRequest struct {
		Name          string              `json:"name" binding:"required"`
		Image         string              `json:"image" binding:"required"`
		Replicas      int32               `json:"replicas" binding:"required"`
		ContainerPort int32               `json:"containerPort"`
		Labels        map[string]string   `json:"labels"`
		Annotations   map[string]string   `json:"annotations"`
		EnvVars       []map[string]string `json:"envVars"`
	}

	if err := c.ShouldBindJSON(&deploymentRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		return
	}

	namespace := c.Param("namespace")

	// Create deployment object
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        deploymentRequest.Name,
			Namespace:   namespace,
			Labels:      deploymentRequest.Labels,
			Annotations: deploymentRequest.Annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &deploymentRequest.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": deploymentRequest.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": deploymentRequest.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  deploymentRequest.Name,
							Image: deploymentRequest.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: deploymentRequest.ContainerPort,
								},
							},
						},
					},
				},
			},
		},
	}

	// Add environment variables if specified
	if len(deploymentRequest.EnvVars) > 0 {
		var envVars []corev1.EnvVar
		for _, env := range deploymentRequest.EnvVars {
			envVars = append(envVars, corev1.EnvVar{
				Name:  env["name"],
				Value: env["value"],
			})
		}
		deployment.Spec.Template.Spec.Containers[0].Env = envVars
	}

	result, err := h.api.CreateDeployment(c.Request.Context(), namespace, deployment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"name":      result.Name,
			"namespace": result.Namespace,
			"status":    "created",
		},
	})
}

// UpdateDeployment handles PUT /api/v1/deployments/namespaces/:namespace/:name
func (h *Handler) UpdateDeployment(c *gin.Context) {
	var updateRequest struct {
		Image       string              `json:"image"`
		Replicas    *int32              `json:"replicas"`
		Labels      map[string]string   `json:"labels"`
		Annotations map[string]string   `json:"annotations"`
		EnvVars     []map[string]string `json:"envVars"`
	}

	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")

	// Get existing deployment
	existing, err := h.api.GetDeployment(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Update fields if provided
	if updateRequest.Image != "" {
		existing.Spec.Template.Spec.Containers[0].Image = updateRequest.Image
	}
	if updateRequest.Replicas != nil {
		existing.Spec.Replicas = updateRequest.Replicas
	}
	if updateRequest.Labels != nil {
		existing.Labels = updateRequest.Labels
	}
	if updateRequest.Annotations != nil {
		existing.Annotations = updateRequest.Annotations
	}
	if len(updateRequest.EnvVars) > 0 {
		var envVars []corev1.EnvVar
		for _, env := range updateRequest.EnvVars {
			envVars = append(envVars, corev1.EnvVar{
				Name:  env["name"],
				Value: env["value"],
			})
		}
		existing.Spec.Template.Spec.Containers[0].Env = envVars
	}

	result, err := h.api.UpdateDeployment(c.Request.Context(), namespace, existing)
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
			"name":      result.Name,
			"namespace": result.Namespace,
			"status":    "updated",
		},
	})
}

func NewHandler(clientset *kubernetes.Clientset, logger *log.Logger) *Handler {
	if logger == nil {
		logger = log.New(gin.DefaultWriter, "[DEPLOYMENT-API] ", log.LstdFlags)
	}

	return &Handler{
		api: NewDeploymentAPI(clientset, logger),
	}
}

// ListDeployments handles GET /api/v1/deployments/namespaces/:namespace
func (h *Handler) ListDeployments(c *gin.Context) {
	namespace := c.Param("namespace")
	deployments, err := h.api.ListDeployments(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	var response []map[string]interface{}
	for _, deployment := range deployments.Items {
		response = append(response, map[string]interface{}{
			"name":          deployment.Name,
			"namespace":     deployment.Namespace,
			"replicas":      deployment.Status.Replicas,
			"readyReplicas": deployment.Status.ReadyReplicas,
			"creationTime":  deployment.CreationTimestamp,
			"labels":        deployment.Labels,
			"strategy":      deployment.Spec.Strategy.Type,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetDeployment handles GET /api/v1/deployments/namespaces/:namespace/:name
func (h *Handler) GetDeployment(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	deployment, err := h.api.GetDeployment(c.Request.Context(), namespace, name)
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
			"name":          deployment.Name,
			"namespace":     deployment.Namespace,
			"replicas":      deployment.Status.Replicas,
			"readyReplicas": deployment.Status.ReadyReplicas,
			"strategy":      deployment.Spec.Strategy.Type,
			"selector":      deployment.Spec.Selector,
			"creationTime":  deployment.CreationTimestamp,
			"labels":        deployment.Labels,
			"annotations":   deployment.Annotations,
			"containers":    deployment.Spec.Template.Spec.Containers,
		},
	})
}

// GetDeploymentStatus handles GET /api/v1/deployments/namespaces/:namespace/:name/status
func (h *Handler) GetDeploymentStatus(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	status, err := h.api.GetDeploymentStatus(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

// DeleteDeployment handles DELETE /api/v1/deployments/namespaces/:namespace/:name
func (h *Handler) DeleteDeployment(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	err := h.api.DeleteDeployment(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Deployment deleted successfully",
	})
}

// ScaleDeployment handles PUT /api/v1/deployments/namespaces/:namespace/:name/scale
func (h *Handler) ScaleDeployment(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	replicasStr := c.Query("replicas")
	replicas, err := strconv.ParseInt(replicasStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid replicas value",
		})
		return
	}

	err = h.api.ScaleDeployment(c.Request.Context(), namespace, name, int32(replicas))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Deployment scaled to %d replicas", replicas),
	})
}

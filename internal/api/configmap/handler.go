package configmap

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Handler struct {
	api *ConfigMapAPI
}

func NewHandler(clientset *kubernetes.Clientset, logger *log.Logger) *Handler {
	if logger == nil {
		logger = log.New(gin.DefaultWriter, "[CONFIGMAP-API] ", log.LstdFlags)
	}

	return &Handler{
		api: NewConfigMapAPI(clientset, logger),
	}
}

// ListConfigMaps handles GET /api/v1/configmaps/namespaces/:namespace
func (h *Handler) ListConfigMaps(c *gin.Context) {
	namespace := c.Param("namespace")
	configMaps, err := h.api.ListConfigMaps(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	var response []map[string]interface{}
	for _, cm := range configMaps.Items {
		response = append(response, map[string]interface{}{
			"name":         cm.Name,
			"namespace":    cm.Namespace,
			"dataCount":    len(cm.Data),
			"creationTime": cm.CreationTimestamp,
			"labels":       cm.Labels,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetConfigMap handles GET /api/v1/configmaps/namespaces/:namespace/:name
func (h *Handler) GetConfigMap(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	configMap, err := h.api.GetConfigMap(c.Request.Context(), namespace, name)
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
			"name":         configMap.Name,
			"namespace":    configMap.Namespace,
			"data":         configMap.Data,
			"binaryData":   configMap.BinaryData,
			"creationTime": configMap.CreationTimestamp,
			"labels":       configMap.Labels,
			"annotations":  configMap.Annotations,
		},
	})
}

// CreateConfigMap handles POST /api/v1/configmaps/namespaces/:namespace
func (h *Handler) CreateConfigMap(c *gin.Context) {
	var configMapRequest struct {
		Name        string            `json:"name" binding:"required"`
		Data        map[string]string `json:"data"`
		BinaryData  map[string][]byte `json:"binaryData"`
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
	}

	if err := c.ShouldBindJSON(&configMapRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		return
	}

	namespace := c.Param("namespace")

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        configMapRequest.Name,
			Namespace:   namespace,
			Labels:      configMapRequest.Labels,
			Annotations: configMapRequest.Annotations,
		},
		Data:       configMapRequest.Data,
		BinaryData: configMapRequest.BinaryData,
	}

	result, err := h.api.CreateConfigMap(c.Request.Context(), namespace, configMap)
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
			"dataCount": len(result.Data),
		},
	})
}

// UpdateConfigMap handles PUT /api/v1/configmaps/namespaces/:namespace/:name
func (h *Handler) UpdateConfigMap(c *gin.Context) {
	var updateRequest struct {
		Data        map[string]string `json:"data"`
		BinaryData  map[string][]byte `json:"binaryData"`
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
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

	// Get existing ConfigMap
	existing, err := h.api.GetConfigMap(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Update fields if provided
	if updateRequest.Data != nil {
		existing.Data = updateRequest.Data
	}
	if updateRequest.BinaryData != nil {
		existing.BinaryData = updateRequest.BinaryData
	}
	if updateRequest.Labels != nil {
		existing.Labels = updateRequest.Labels
	}
	if updateRequest.Annotations != nil {
		existing.Annotations = updateRequest.Annotations
	}

	result, err := h.api.UpdateConfigMap(c.Request.Context(), namespace, existing)
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
			"dataCount": len(result.Data),
			"status":    "updated",
		},
	})
}

// DeleteConfigMap handles DELETE /api/v1/configmaps/namespaces/:namespace/:name
func (h *Handler) DeleteConfigMap(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	err := h.api.DeleteConfigMap(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "ConfigMap deleted successfully",
	})
}

// GetConfigMapUsage handles GET /api/v1/configmaps/namespaces/:namespace/:name/usage
func (h *Handler) GetConfigMapUsage(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	usage, err := h.api.GetConfigMapUsage(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, usage)
}

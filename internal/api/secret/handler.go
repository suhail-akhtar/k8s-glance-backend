package secret

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Handler struct {
	api *SecretAPI
}

func NewHandler(clientset *kubernetes.Clientset, logger *log.Logger) *Handler {
	if logger == nil {
		logger = log.New(gin.DefaultWriter, "[SECRET-API] ", log.LstdFlags)
	}

	return &Handler{
		api: NewSecretAPI(clientset, logger),
	}
}

// ListSecrets handles GET /api/v1/secrets/namespaces/:namespace
func (h *Handler) ListSecrets(c *gin.Context) {
	namespace := c.Param("namespace")
	secrets, err := h.api.ListSecrets(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	var response []map[string]interface{}
	for _, secret := range secrets.Items {
		response = append(response, map[string]interface{}{
			"name":         secret.Name,
			"namespace":    secret.Namespace,
			"type":         string(secret.Type),
			"creationTime": secret.CreationTimestamp,
			"labels":       secret.Labels,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetSecret handles GET /api/v1/secrets/namespaces/:namespace/:name
func (h *Handler) GetSecret(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	secret, err := h.api.GetSecret(c.Request.Context(), namespace, name)
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
			"name":         secret.Name,
			"namespace":    secret.Namespace,
			"type":         string(secret.Type),
			"creationTime": secret.CreationTimestamp,
			"labels":       secret.Labels,
			"annotations":  secret.Annotations,
		},
	})
}

// GetSecretKeys handles GET /api/v1/secrets/namespaces/:namespace/:name/keys
func (h *Handler) GetSecretKeys(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	keys, err := h.api.GetSecretKeys(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, keys)
}

// CreateSecret handles POST /api/v1/secrets/namespaces/:namespace
func (h *Handler) CreateSecret(c *gin.Context) {
	var secretRequest struct {
		Name        string            `json:"name" binding:"required"`
		Type        string            `json:"type"`
		StringData  map[string]string `json:"stringData"`
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
	}

	if err := c.ShouldBindJSON(&secretRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		return
	}

	namespace := c.Param("namespace")

	// Set default type if not provided
	if secretRequest.Type == "" {
		secretRequest.Type = "Opaque"
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        secretRequest.Name,
			Namespace:   namespace,
			Labels:      secretRequest.Labels,
			Annotations: secretRequest.Annotations,
		},
		Type:       corev1.SecretType(secretRequest.Type),
		StringData: secretRequest.StringData,
	}

	result, err := h.api.CreateSecret(c.Request.Context(), namespace, secret)
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
			"type":      string(result.Type),
		},
	})
}

// UpdateSecret handles PUT /api/v1/secrets/namespaces/:namespace/:name
func (h *Handler) UpdateSecret(c *gin.Context) {
	var updateRequest struct {
		StringData  map[string]string `json:"stringData"`
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

	// Get existing secret
	existing, err := h.api.GetSecret(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Update fields if provided
	if updateRequest.StringData != nil {
		// Convert string data to Secret Data
		if existing.Data == nil {
			existing.Data = make(map[string][]byte)
		}
		for k, v := range updateRequest.StringData {
			existing.Data[k] = []byte(v)
		}
	}

	if updateRequest.Labels != nil {
		existing.Labels = updateRequest.Labels
	}
	if updateRequest.Annotations != nil {
		existing.Annotations = updateRequest.Annotations
	}

	result, err := h.api.UpdateSecret(c.Request.Context(), namespace, existing)
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
			"type":      string(result.Type),
			"status":    "updated",
		},
	})
}

// DeleteSecret handles DELETE /api/v1/secrets/namespaces/:namespace/:name
func (h *Handler) DeleteSecret(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	err := h.api.DeleteSecret(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Secret deleted successfully",
	})
}

// GetSecretUsage handles GET /api/v1/secrets/namespaces/:namespace/:name/usage
func (h *Handler) GetSecretUsage(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	usage, err := h.api.GetSecretUsage(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, usage)
}

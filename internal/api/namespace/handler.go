package namespace

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes"
)

type Handler struct {
	api *NamespaceAPI
}

func NewHandler(clientset *kubernetes.Clientset, logger *log.Logger) *Handler {
	if logger == nil {
		logger = log.New(gin.DefaultWriter, "[NAMESPACE-API] ", log.LstdFlags)
	}

	return &Handler{
		api: NewNamespaceAPI(clientset, logger),
	}
}

// ListNamespaces handles GET /api/v1/namespaces
func (h *Handler) ListNamespaces(c *gin.Context) {
	namespaces, err := h.api.ListNamespaces(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Convert to a simpler response format
	var response []map[string]interface{}
	for _, ns := range namespaces.Items {
		response = append(response, map[string]interface{}{
			"name":            ns.Name,
			"status":          ns.Status.Phase,
			"creationTime":    ns.CreationTimestamp,
			"resourceVersion": ns.ResourceVersion,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetNamespace handles GET /api/v1/namespaces/:name
func (h *Handler) GetNamespace(c *gin.Context) {
	name := c.Param("name")
	namespace, err := h.api.GetNamespace(c.Request.Context(), name)

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
			"name":            namespace.Name,
			"status":          namespace.Status.Phase,
			"creationTime":    namespace.CreationTimestamp,
			"resourceVersion": namespace.ResourceVersion,
			"labels":          namespace.Labels,
			"annotations":     namespace.Annotations,
		},
	})
}

// GetNamespaceMetrics handles GET /api/v1/namespaces/:name/metrics
func (h *Handler) GetNamespaceMetrics(c *gin.Context) {
	name := c.Param("name")
	metrics, err := h.api.GetNamespaceMetrics(c.Request.Context(), name)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

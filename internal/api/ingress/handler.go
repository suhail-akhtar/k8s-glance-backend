package ingress

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Handler struct {
	api *IngressAPI
}

// Helper functions
func getTLSConfig(tls []networkingv1.IngressTLS) []map[string]interface{} {
	var result []map[string]interface{}
	for _, t := range tls {
		result = append(result, map[string]interface{}{
			"hosts":      t.Hosts,
			"secretName": t.SecretName,
		})
	}
	return result
}

func NewHandler(clientset *kubernetes.Clientset, logger *log.Logger) *Handler {
	if logger == nil {
		logger = log.New(gin.DefaultWriter, "[INGRESS-API] ", log.LstdFlags)
	}

	return &Handler{
		api: NewIngressAPI(clientset, logger),
	}
}

// ListIngresses handles GET /api/v1/namespaces/:namespace/ingresses
func (h *Handler) ListIngresses(c *gin.Context) {
	namespace := c.Param("namespace")
	ingresses, err := h.api.ListIngresses(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	var response []map[string]interface{}
	for _, ing := range ingresses.Items {
		response = append(response, map[string]interface{}{
			"name":         ing.Name,
			"namespace":    ing.Namespace,
			"className":    ing.Spec.IngressClassName,
			"rules":        getIngressRules(ing.Spec.Rules),
			"creationTime": ing.CreationTimestamp,
			"labels":       ing.Labels,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// CreateIngress handles POST /api/v1/namespaces/:namespace/ingresses
func (h *Handler) CreateIngress(c *gin.Context) {
	var ingressRequest struct {
		Name      string `json:"name" binding:"required"`
		ClassName string `json:"className"`
		Rules     []struct {
			Host  string `json:"host"`
			Paths []struct {
				Path        string `json:"path" binding:"required"`
				PathType    string `json:"pathType" binding:"required"`
				ServiceName string `json:"serviceName" binding:"required"`
				ServicePort int32  `json:"servicePort" binding:"required"`
			} `json:"paths" binding:"required"`
		} `json:"rules" binding:"required"`
		TLS []struct {
			Hosts      []string `json:"hosts"`
			SecretName string   `json:"secretName"`
		} `json:"tls"`
		Annotations map[string]string `json:"annotations"`
		Labels      map[string]string `json:"labels"`
	}

	if err := c.ShouldBindJSON(&ingressRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		return
	}

	namespace := c.Param("namespace")

	// Create ingress rules
	var rules []networkingv1.IngressRule
	for _, rule := range ingressRequest.Rules {
		var paths []networkingv1.HTTPIngressPath
		for _, path := range rule.Paths {
			pathType := networkingv1.PathType(path.PathType)
			paths = append(paths, networkingv1.HTTPIngressPath{
				Path:     path.Path,
				PathType: &pathType,
				Backend: networkingv1.IngressBackend{
					Service: &networkingv1.IngressServiceBackend{
						Name: path.ServiceName,
						Port: networkingv1.ServiceBackendPort{
							Number: path.ServicePort,
						},
					},
				},
			})
		}

		rules = append(rules, networkingv1.IngressRule{
			Host: rule.Host,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: paths,
				},
			},
		})
	}

	// Create TLS configuration
	var tls []networkingv1.IngressTLS
	for _, t := range ingressRequest.TLS {
		tls = append(tls, networkingv1.IngressTLS{
			Hosts:      t.Hosts,
			SecretName: t.SecretName,
		})
	}

	// Create ingress object
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        ingressRequest.Name,
			Namespace:   namespace,
			Labels:      ingressRequest.Labels,
			Annotations: ingressRequest.Annotations,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &ingressRequest.ClassName,
			Rules:            rules,
			TLS:              tls,
		},
	}

	result, err := h.api.CreateIngress(c.Request.Context(), namespace, ingress)
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
			"className": result.Spec.IngressClassName,
			"rules":     getIngressRules(result.Spec.Rules),
		},
	})
}

// UpdateIngress handles PUT /api/v1/namespaces/:namespace/ingresses/:name
func (h *Handler) UpdateIngress(c *gin.Context) {
	var updateRequest struct {
		ClassName string `json:"className"`
		Rules     []struct {
			Host  string `json:"host"`
			Paths []struct {
				Path        string `json:"path"`
				PathType    string `json:"pathType"`
				ServiceName string `json:"serviceName"`
				ServicePort int32  `json:"servicePort"`
			} `json:"paths"`
		} `json:"rules"`
		TLS []struct {
			Hosts      []string `json:"hosts"`
			SecretName string   `json:"secretName"`
		} `json:"tls"`
		Annotations map[string]string `json:"annotations"`
		Labels      map[string]string `json:"labels"`
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

	// Get existing ingress
	existing, err := h.api.GetIngress(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Update fields if provided
	if updateRequest.ClassName != "" {
		existing.Spec.IngressClassName = &updateRequest.ClassName
	}

	if len(updateRequest.Rules) > 0 {
		var rules []networkingv1.IngressRule
		for _, rule := range updateRequest.Rules {
			var paths []networkingv1.HTTPIngressPath
			for _, path := range rule.Paths {
				pathType := networkingv1.PathType(path.PathType)
				paths = append(paths, networkingv1.HTTPIngressPath{
					Path:     path.Path,
					PathType: &pathType,
					Backend: networkingv1.IngressBackend{
						Service: &networkingv1.IngressServiceBackend{
							Name: path.ServiceName,
							Port: networkingv1.ServiceBackendPort{
								Number: path.ServicePort,
							},
						},
					},
				},
				)
			}

			rules = append(rules, networkingv1.IngressRule{
				Host: rule.Host,
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: paths,
					},
				},
			})
		}
		existing.Spec.Rules = rules
	}

	if len(updateRequest.TLS) > 0 {
		var tls []networkingv1.IngressTLS
		for _, t := range updateRequest.TLS {
			tls = append(tls, networkingv1.IngressTLS{
				Hosts:      t.Hosts,
				SecretName: t.SecretName,
			})
		}
		existing.Spec.TLS = tls
	}

	if updateRequest.Labels != nil {
		existing.Labels = updateRequest.Labels
	}

	if updateRequest.Annotations != nil {
		existing.Annotations = updateRequest.Annotations
	}

	result, err := h.api.UpdateIngress(c.Request.Context(), namespace, existing)
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
			"className": result.Spec.IngressClassName,
			"rules":     getIngressRules(result.Spec.Rules),
			"status":    "updated",
		},
	})
}

// GetIngress handles GET /api/v1/namespaces/:namespace/ingresses/:name
func (h *Handler) GetIngress(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	ingress, err := h.api.GetIngress(c.Request.Context(), namespace, name)
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
			"name":         ingress.Name,
			"namespace":    ingress.Namespace,
			"className":    ingress.Spec.IngressClassName,
			"rules":        getIngressRules(ingress.Spec.Rules),
			"tls":          getTLSConfig(ingress.Spec.TLS),
			"creationTime": ingress.CreationTimestamp,
			"labels":       ingress.Labels,
			"annotations":  ingress.Annotations,
		},
	})
}

// DeleteIngress handles DELETE /api/v1/namespaces/:namespace/ingresses/:name
func (h *Handler) DeleteIngress(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	err := h.api.DeleteIngress(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Ingress deleted successfully",
	})
}

// GetIngressStatus handles GET /api/v1/namespaces/:namespace/ingresses/:name/status
func (h *Handler) GetIngressStatus(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	status, err := h.api.GetIngressStatus(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

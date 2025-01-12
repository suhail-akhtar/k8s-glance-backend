package service

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

type Handler struct {
	api *ServiceAPI
}

func NewHandler(clientset *kubernetes.Clientset, logger *log.Logger) *Handler {
	if logger == nil {
		logger = log.New(gin.DefaultWriter, "[SERVICE-API] ", log.LstdFlags)
	}

	return &Handler{
		api: NewServiceAPI(clientset, logger),
	}
}

// ListServices handles GET /api/v1/services/namespaces/:namespace
func (h *Handler) ListServices(c *gin.Context) {
	namespace := c.Param("namespace")
	services, err := h.api.ListServices(c.Request.Context(), namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	var response []map[string]interface{}
	for _, svc := range services.Items {
		response = append(response, map[string]interface{}{
			"name":      svc.Name,
			"namespace": svc.Namespace,
			"type":      string(svc.Spec.Type),
			"clusterIP": svc.Spec.ClusterIP,
			"ports":     svc.Spec.Ports,
			"selector":  svc.Spec.Selector,
			"created":   svc.CreationTimestamp,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetService handles GET /api/v1/services/namespaces/:namespace/:name
func (h *Handler) GetService(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	service, err := h.api.GetService(c.Request.Context(), namespace, name)
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
			"name":            service.Name,
			"namespace":       service.Namespace,
			"type":            string(service.Spec.Type),
			"clusterIP":       service.Spec.ClusterIP,
			"ports":           service.Spec.Ports,
			"selector":        service.Spec.Selector,
			"sessionAffinity": string(service.Spec.SessionAffinity),
			"created":         service.CreationTimestamp,
			"labels":          service.Labels,
			"annotations":     service.Annotations,
		},
	})
}

// CreateService handles POST /api/v1/services/namespaces/:namespace
func (h *Handler) CreateService(c *gin.Context) {
	var serviceRequest struct {
		Name  string `json:"name" binding:"required"`
		Type  string `json:"type" binding:"required"`
		Ports []struct {
			Name       string `json:"name"`
			Port       int32  `json:"port" binding:"required"`
			TargetPort int32  `json:"targetPort"`
			NodePort   int32  `json:"nodePort,omitempty"`
			Protocol   string `json:"protocol,omitempty"`
		} `json:"ports" binding:"required"`
		Selector    map[string]string `json:"selector" binding:"required"`
		Labels      map[string]string `json:"labels"`
		ExternalIPs []string          `json:"externalIPs"`
	}

	if err := c.ShouldBindJSON(&serviceRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		return
	}

	namespace := c.Param("namespace")

	// Create service ports
	var servicePorts []corev1.ServicePort
	for _, p := range serviceRequest.Ports {
		protocol := corev1.ProtocolTCP
		if p.Protocol != "" {
			protocol = corev1.Protocol(p.Protocol)
		}

		servicePort := corev1.ServicePort{
			Name:       p.Name,
			Port:       p.Port,
			TargetPort: intstr.FromInt(int(p.TargetPort)),
			Protocol:   protocol,
		}

		if p.NodePort > 0 {
			servicePort.NodePort = p.NodePort
		}

		servicePorts = append(servicePorts, servicePort)
	}

	// Create service object
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceRequest.Name,
			Namespace: namespace,
			Labels:    serviceRequest.Labels,
		},
		Spec: corev1.ServiceSpec{
			Type:        corev1.ServiceType(serviceRequest.Type),
			Ports:       servicePorts,
			Selector:    serviceRequest.Selector,
			ExternalIPs: serviceRequest.ExternalIPs,
		},
	}

	result, err := h.api.CreateService(c.Request.Context(), namespace, service)
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
			"type":      string(result.Spec.Type),
			"clusterIP": result.Spec.ClusterIP,
		},
	})
}

// UpdateService handles PUT /api/v1/services/namespaces/:namespace/:name
func (h *Handler) UpdateService(c *gin.Context) {
	var updateRequest struct {
		Ports []struct {
			Name       string `json:"name"`
			Port       int32  `json:"port"`
			TargetPort int32  `json:"targetPort"`
			NodePort   int32  `json:"nodePort,omitempty"`
			Protocol   string `json:"protocol,omitempty"`
		} `json:"ports"`
		Selector    map[string]string `json:"selector"`
		Labels      map[string]string `json:"labels"`
		ExternalIPs []string          `json:"externalIPs"`
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

	// Get existing service
	existing, err := h.api.GetService(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Update fields if provided
	if len(updateRequest.Ports) > 0 {
		var servicePorts []corev1.ServicePort
		for _, p := range updateRequest.Ports {
			protocol := corev1.ProtocolTCP
			if p.Protocol != "" {
				protocol = corev1.Protocol(p.Protocol)
			}

			servicePort := corev1.ServicePort{
				Name:       p.Name,
				Port:       p.Port,
				TargetPort: intstr.FromInt(int(p.TargetPort)),
				Protocol:   protocol,
			}

			if p.NodePort > 0 {
				servicePort.NodePort = p.NodePort
			}

			servicePorts = append(servicePorts, servicePort)
		}
		existing.Spec.Ports = servicePorts
	}

	if updateRequest.Selector != nil {
		existing.Spec.Selector = updateRequest.Selector
	}

	if updateRequest.Labels != nil {
		existing.Labels = updateRequest.Labels
	}

	if updateRequest.ExternalIPs != nil {
		existing.Spec.ExternalIPs = updateRequest.ExternalIPs
	}

	result, err := h.api.UpdateService(c.Request.Context(), namespace, existing)
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
			"type":      string(result.Spec.Type),
			"clusterIP": result.Spec.ClusterIP,
			"status":    "updated",
		},
	})
}

// DeleteService handles DELETE /api/v1/services/namespaces/:namespace/:name
func (h *Handler) DeleteService(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	err := h.api.DeleteService(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Service deleted successfully",
	})
}

// GetServiceStatus handles GET /api/v1/services/namespaces/:namespace/:name/status
func (h *Handler) GetServiceStatus(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	status, err := h.api.GetServiceStatus(c.Request.Context(), namespace, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, status)
}

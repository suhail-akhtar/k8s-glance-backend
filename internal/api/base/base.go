package base

import (
	"context"
	"fmt"
	"log"
	"os"

	errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// BaseAPI provides common functionality for all API services
type BaseAPI struct {
	clientset *kubernetes.Clientset
	logger    *log.Logger
}

// NewBaseAPI creates a new instance of BaseAPI
func NewBaseAPI(clientset *kubernetes.Clientset, logger *log.Logger) *BaseAPI {
	if logger == nil {
		logger = log.New(os.Stdout, "[BASE-API] ", log.LstdFlags)
	}

	return &BaseAPI{
		clientset: clientset,
		logger:    logger,
	}
}

// GetClientset returns the kubernetes clientset
func (b *BaseAPI) GetClientset() *kubernetes.Clientset {
	return b.clientset
}

// LogError logs an error with context
func (b *BaseAPI) LogError(ctx context.Context, operation string, err error) {
	if statusErr, ok := err.(*errors.StatusError); ok {
		b.logger.Printf("Error during %s: %v (Code: %d, Reason: %s, Details: %+v)",
			operation, err, statusErr.ErrStatus.Code, statusErr.ErrStatus.Reason, statusErr.ErrStatus.Details)
	} else {
		b.logger.Printf("Error during %s: %v", operation, err)
	}
}

// LogInfo logs information with context
func (b *BaseAPI) LogInfo(ctx context.Context, operation string, message string) {
	b.logger.Printf("Info [%s]: %s", operation, message)
}

// HandleError standardizes error handling across APIs
func (b *BaseAPI) HandleError(err error, operation string) error {
	if err != nil {
		b.LogError(context.Background(), operation, err)
		if statusErr, ok := err.(*errors.StatusError); ok {
			return fmt.Errorf("%s failed: %v (Code: %d, Reason: %s)",
				operation, err, statusErr.ErrStatus.Code, statusErr.ErrStatus.Reason)
		}
		return fmt.Errorf("%s failed: %w", operation, err)
	}
	return nil
}

// IsHealthy checks if the kubernetes API is accessible
func (b *BaseAPI) IsHealthy(ctx context.Context) error {
	_, err := b.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
	return b.HandleError(err, "health check")
}

// Common response structures
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// NewSuccessResponse creates a success response
func NewSuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Data:    data,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(err error) APIResponse {
	return APIResponse{
		Success: false,
		Error:   err.Error(),
	}
}

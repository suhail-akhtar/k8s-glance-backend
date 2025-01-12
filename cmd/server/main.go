package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes" // Added this import

	"k8s-glance-backend/internal/api/configmap"
	"k8s-glance-backend/internal/api/deployment"
	"k8s-glance-backend/internal/api/ingress"
	"k8s-glance-backend/internal/api/namespace"
	"k8s-glance-backend/internal/api/pod"
	"k8s-glance-backend/internal/api/secret"
	"k8s-glance-backend/internal/api/service"
	"k8s-glance-backend/internal/config"
	k8sclient "k8s-glance-backend/pkg/kubernetes" // Aliased to avoid confusion
)

func main() {
	// Initialize logger
	logger := log.New(os.Stdout, "[K8S-GLANCE] ", log.LstdFlags)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Kubernetes client
	k8sClient, err := k8sclient.NewClient(cfg.KubeConfig)
	if err != nil {
		logger.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Initialize Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Setup routes
	setupRoutes(router, k8sClient.Clientset, logger)

	// Create server with timeout configurations
	srv := &http.Server{
		Addr:         cfg.ServerAddress,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Printf("Starting server on %s", cfg.ServerAddress)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	logger.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Println("Server exiting")
}

func setupRoutes(router *gin.Engine, clientset *kubernetes.Clientset, logger *log.Logger) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Initialize handlers
	namespaceHandler := namespace.NewHandler(clientset, logger)
	podHandler := pod.NewHandler(clientset, logger)
	deploymentHandler := deployment.NewHandler(clientset, logger)
	serviceHandler := service.NewHandler(clientset, logger)
	configMapHandler := configmap.NewHandler(clientset, logger)
	secretHandler := secret.NewHandler(clientset, logger)
	ingressHandler := ingress.NewHandler(clientset, logger)

	// API version group
	v1 := router.Group("/api/v1")
	{
		// Namespace routes
		namespaces := v1.Group("/namespaces")
		{
			namespaces.GET("", namespaceHandler.ListNamespaces)
			namespaces.GET("/:name", namespaceHandler.GetNamespace)
			namespaces.GET("/:name/metrics", namespaceHandler.GetNamespaceMetrics)
		}

		// Pod routes
		pods := v1.Group("/pods")
		{
			pods.GET("/namespaces/:namespace", podHandler.ListPods)
			pods.GET("/namespaces/:namespace/:name", podHandler.GetPod)
			pods.GET("/namespaces/:namespace/:name/metrics", podHandler.GetPodMetrics)
			pods.DELETE("/namespaces/:namespace/:name", podHandler.DeletePod)
		}

		// Deployment routes
		deployments := v1.Group("/deployments")
		{
			deployments.GET("/namespaces/:namespace", deploymentHandler.ListDeployments)
			deployments.POST("/namespaces/:namespace", deploymentHandler.CreateDeployment)
			deployments.GET("/namespaces/:namespace/:name", deploymentHandler.GetDeployment)
			deployments.PUT("/namespaces/:namespace/:name", deploymentHandler.UpdateDeployment)
			deployments.GET("/namespaces/:namespace/:name/status", deploymentHandler.GetDeploymentStatus)
			deployments.DELETE("/namespaces/:namespace/:name", deploymentHandler.DeleteDeployment)
			deployments.PUT("/namespaces/:namespace/:name/scale", deploymentHandler.ScaleDeployment)
		}

		// Service routes
		services := v1.Group("/services")
		{
			services.GET("/namespaces/:namespace", serviceHandler.ListServices)
			services.POST("/namespaces/:namespace", serviceHandler.CreateService)
			services.GET("/namespaces/:namespace/:name", serviceHandler.GetService)
			services.PUT("/namespaces/:namespace/:name", serviceHandler.UpdateService)
			services.DELETE("/namespaces/:namespace/:name", serviceHandler.DeleteService)
			services.GET("/namespaces/:namespace/:name/status", serviceHandler.GetServiceStatus)
		}

		// ConfigMap routes
		configMaps := v1.Group("/configmaps")
		{
			configMaps.GET("/namespaces/:namespace", configMapHandler.ListConfigMaps)
			configMaps.POST("/namespaces/:namespace", configMapHandler.CreateConfigMap)
			configMaps.GET("/namespaces/:namespace/:name", configMapHandler.GetConfigMap)
			configMaps.PUT("/namespaces/:namespace/:name", configMapHandler.UpdateConfigMap)
			configMaps.DELETE("/namespaces/:namespace/:name", configMapHandler.DeleteConfigMap)
			configMaps.GET("/namespaces/:namespace/:name/usage", configMapHandler.GetConfigMapUsage)
		}
		// Secret routes
		secrets := v1.Group("/secrets")
		{
			secrets.GET("/namespaces/:namespace", secretHandler.ListSecrets)
			secrets.POST("/namespaces/:namespace", secretHandler.CreateSecret)
			secrets.GET("/namespaces/:namespace/:name", secretHandler.GetSecret)
			secrets.PUT("/namespaces/:namespace/:name", secretHandler.UpdateSecret)
			secrets.DELETE("/namespaces/:namespace/:name", secretHandler.DeleteSecret)
		}

		// Ingress routes (nested under namespaces)
		ingresses := namespaces.Group("/:namespace/ingresses")
		{
			ingresses.GET("", ingressHandler.ListIngresses)
			ingresses.POST("", ingressHandler.CreateIngress)
			ingresses.GET("/:name", ingressHandler.GetIngress)
			ingresses.PUT("/:name", ingressHandler.UpdateIngress)
			ingresses.DELETE("/:name", ingressHandler.DeleteIngress)
			ingresses.GET("/:name/status", ingressHandler.GetIngressStatus)
		}
	}
}

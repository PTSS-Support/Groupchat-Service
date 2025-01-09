package config

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// ServiceDiscoveryConfig holds configuration for service discovery
type ServiceDiscoveryConfig struct {
	// The OpenShift namespace where services are deployed
	Namespace string `mapstructure:"namespace"`
	// The service name to look for
	ServiceName string `mapstructure:"service_name"`
	// The port name or number to connect to
	PortName string `mapstructure:"port_name"`
	// How often to refresh service discovery (in seconds)
	RefreshInterval int `mapstructure:"refresh_interval"`
}

// ServiceDiscoverer handles OpenShift service discovery
type ServiceDiscoverer struct {
	config     *ServiceDiscoveryConfig
	kubeClient *kubernetes.Clientset
}

// NewServiceDiscoverer creates a new service discoverer
func NewServiceDiscoverer(config *ServiceDiscoveryConfig) (*ServiceDiscoverer, error) {
	// Get in-cluster config
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	// Create kubernetes client
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &ServiceDiscoverer{
		config:     config,
		kubeClient: kubeClient,
	}, nil
}

// DiscoverServiceURL discovers the service URL and returns it
func (sd *ServiceDiscoverer) DiscoverServiceURL(ctx context.Context) (string, error) {
	svc, err := sd.kubeClient.CoreV1().Services(sd.config.Namespace).Get(ctx, sd.config.ServiceName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to discover service %s: %w", sd.config.ServiceName, err)
	}

	// Find the specified port
	var port int32
	for _, p := range svc.Spec.Ports {
		if p.Name == sd.config.PortName || fmt.Sprint(p.Port) == sd.config.PortName {
			port = p.Port
			break
		}
	}

	if port == 0 {
		return "", fmt.Errorf("port %s not found in service %s", sd.config.PortName, sd.config.ServiceName)
	}

	// Construct service URL using OpenShift DNS naming convention
	serviceURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
		sd.config.ServiceName,
		sd.config.Namespace,
		port)

	return serviceURL, nil
}

// StartDiscovery starts the service discovery process
func (sd *ServiceDiscoverer) StartDiscovery(ctx context.Context) (<-chan string, <-chan error) {
	urls := make(chan string)
	errs := make(chan error)

	go func() {
		defer close(urls)
		defer close(errs)

		ticker := time.NewTicker(time.Duration(sd.config.RefreshInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				url, err := sd.DiscoverServiceURL(ctx)
				if err != nil {
					errs <- err
					continue
				}
				urls <- url
			}
		}
	}()

	return urls, errs
}

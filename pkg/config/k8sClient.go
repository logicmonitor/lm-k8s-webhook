package config

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// K8sClient represents the Kubernetes client object
type K8sClient struct {
	Clientset kubernetes.Interface
}

// NewK8sClient creates and returns kuberentes client
func NewK8sClient(k8sRestConfig *rest.Config, k8sClientSet func(*rest.Config) (kubernetes.Interface, error)) (*K8sClient, error) {
	clientSet, err := k8sClientSet(k8sRestConfig)
	if err != nil {
		return nil, err
	}
	return &K8sClient{Clientset: clientSet}, nil
}

// NewK8sClientSet creates and returns client set
func NewK8sClientSet(k8sRestConfig *rest.Config) (kubernetes.Interface, error) {
	clientset, err := kubernetes.NewForConfig(k8sRestConfig)
	if err != nil {
		return clientset, err
	}
	return clientset, nil
}

package config

import (
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sClient represents the Kubernetes client object
type K8sClient struct {
	Clientset kubernetes.Interface
}

// K8sRestConfig gets rest config for K8s cluster
func K8sRestConfig() (*rest.Config, error) {

	var config *rest.Config

	var err error

	kubeconfigPath := os.Getenv("KUBECONFIG")

	if kubeconfigPath == "" {
		kubeconfigPath = os.Getenv("HOME") + "/.kube/config"
	}

	//If file exists so use that config settings
	if _, err = os.Stat(kubeconfigPath); err == nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, err
		}
	} else { //Use Incluster Configuration
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}
	return config, nil
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

package mutation

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/logicmonitor/lm-k8s-webhook/pkg/config"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	testclient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var logger = logf.Log.WithName("unit-tests")

// getFakeK8sClient returns the dummy kubernetes client object for testing
func getFakeK8sClient() (*config.K8sClient, error) {
	return config.NewK8sClient(nil, func(r *rest.Config) (kubernetes.Interface, error) {
		return testclient.NewSimpleClientset(
			//	&batchv1.Job{ObjectMeta: v1.ObjectMeta{Name: "hello-job", Namespace: "default"}},
			//	&appsv1.DaemonSet{ObjectMeta: v1.ObjectMeta{Name: "hello-daemonSet", Namespace: "default"}},
			//	&appsv1.StatefulSet{ObjectMeta: v1.ObjectMeta{Name: "hello-statefulSet", Namespace: "default"}},
			&appsv1.ReplicaSet{ObjectMeta: v1.ObjectMeta{Name: "hello-replicaSet", Namespace: "default"}},
			&appsv1.ReplicaSet{ObjectMeta: v1.ObjectMeta{Name: "hello-replicaSetManagedByDeployment", Namespace: "default", OwnerReferences: []v1.OwnerReference{{Name: "hello-deployment", Kind: "Deployment"}}}},
		), nil
	})
}

func TestMutatePod(t *testing.T) {

	cmpOpt := cmp.AllowUnexported()
	k8sClient, err := getFakeK8sClient()
	if err != nil {
		t.Errorf("Error occured in getting fake k8s client: %v", err)
		return
	}
	os.Setenv("CLUSTER_NAME", "default")
	defer os.Unsetenv("CLUSTER_NAME")

	tests := []struct {
		name string
		args struct {
			params *Params
			ctx    context.Context
		}
		wantErr     bool
		wantPayload corev1.Pod
	}{
		{
			name: "Mutate pod without external config",
			args: struct {
				params *Params
				ctx    context.Context
			}{
				params: &Params{
					Client:    k8sClient,
					LMConfig:  config.Config{},
					Log:       logger,
					Namespace: "default",
					Pod: &corev1.Pod{
						ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "my-app",
									Env: []corev1.EnvVar{
										{Name: "DEPARTMENT",
											Value: "R&D",
										},
									},
								},
							},
						},
					},
				},
				ctx: context.Background(),
			},
			wantErr: false,
			wantPayload: corev1.Pod{
				ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "my-app",
							Env: []corev1.EnvVar{
								{
									Name:  LMAPMClusterName,
									Value: os.Getenv(ClusterName),
								},
								{
									Name:      LMAPMNodeName,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}},
								},
								{
									Name:      LMAPMPodName,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
								},
								{
									Name:      LMAPMPodNamespace,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
								},
								{
									Name:      LMAPMPodIP,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}},
								},
								{
									Name:      LMAPMPodUID,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.uid"}},
								},
								{
									Name:      ServiceNamespace,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
								},
								{
									Name:  ServiceName,
									Value: "test-pod",
								},
								{
									Name:  "DEPARTMENT",
									Value: "R&D",
								},
								{
									Name:  "OTEL_RESOURCE_ATTRIBUTES",
									Value: "host.name=$(LM_APM_POD_NAME),ip=$(LM_APM_POD_IP),k8s.cluster.name=$(LM_APM_CLUSTER_NAME),k8s.namespace.name=$(LM_APM_POD_NAMESPACE),k8s.node.name=$(LM_APM_NODE_NAME),k8s.pod.uid=$(LM_APM_POD_UID),resource.type=kubernetes-pod,service.namespace=$(SERVICE_NAMESPACE),service.name=$(SERVICE_NAME)",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Mutate pod without external config but SERVICE_NAMESPACE & SERVICE_NAME present in container already",
			args: struct {
				params *Params
				ctx    context.Context
			}{
				params: &Params{
					Client:    k8sClient,
					LMConfig:  config.Config{},
					Log:       logger,
					Namespace: "default",
					Pod: &corev1.Pod{
						ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "my-app",
									Env: []corev1.EnvVar{
										{
											Name:  "DEPARTMENT",
											Value: "R&D",
										},
										{
											Name:  "SERVICE_NAMESPACE",
											Value: "hipster",
										},
										{
											Name:  "SERVICE_NAME",
											Value: "test",
										},
									},
								},
							},
						},
					},
				},
				ctx: context.Background(),
			},
			wantErr: false,
			wantPayload: corev1.Pod{
				ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "my-app",
							Env: []corev1.EnvVar{
								{
									Name:  LMAPMClusterName,
									Value: os.Getenv(ClusterName),
								},
								{
									Name:      LMAPMNodeName,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}},
								},
								{
									Name:      LMAPMPodName,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
								},
								{
									Name:      LMAPMPodNamespace,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
								},
								{
									Name:      LMAPMPodIP,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}},
								},
								{
									Name:      LMAPMPodUID,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.uid"}},
								},
								{
									Name:  ServiceNamespace,
									Value: "hipster",
								},
								{
									Name:  ServiceName,
									Value: "test",
								},
								{
									Name:  "DEPARTMENT",
									Value: "R&D",
								},
								{
									Name:  "OTEL_RESOURCE_ATTRIBUTES",
									Value: "host.name=$(LM_APM_POD_NAME),ip=$(LM_APM_POD_IP),k8s.cluster.name=$(LM_APM_CLUSTER_NAME),k8s.namespace.name=$(LM_APM_POD_NAMESPACE),k8s.node.name=$(LM_APM_NODE_NAME),k8s.pod.uid=$(LM_APM_POD_UID),resource.type=kubernetes-pod,service.namespace=$(SERVICE_NAMESPACE),service.name=$(SERVICE_NAME)",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Mutate pod with external config",
			args: struct {
				params *Params
				ctx    context.Context
			}{
				params: &Params{
					Client: k8sClient,
					Log:    logger,
					LMConfig: config.Config{
						MutationConfigProvided: true,
						MutationConfig: config.MutationConfig{
							LMEnvVars: config.LMEnvVars{
								Resource: []config.ResourceEnv{
									{
										Env: corev1.EnvVar{
											Name: "SERVICE_NAMESPACE",
											ValueFrom: &corev1.EnvVarSource{
												FieldRef: &corev1.ObjectFieldSelector{
													FieldPath: "metadata.labels['app-namespace']",
												},
											},
										},
									},
									{
										Env: corev1.EnvVar{
											Name: "SERVICE_NAME",
											ValueFrom: &corev1.EnvVarSource{
												FieldRef: &corev1.ObjectFieldSelector{
													FieldPath: "metadata.labels['app-name']",
												},
											},
										},
									},
									{
										Env: corev1.EnvVar{
											Name: "SERVICE_ACCOUNT_NAME",
											ValueFrom: &corev1.EnvVarSource{
												FieldRef: &corev1.ObjectFieldSelector{
													FieldPath: "spec.serviceAccountName",
												},
											},
										},
									},
								},
								Operation: []config.OperationEnv{
									{
										Env: corev1.EnvVar{
											Name:  "OTLP_ENDPOINT",
											Value: "lmotel-svc:4317",
										},
									},
								},
							},
						},
					},
					Pod: &corev1.Pod{
						ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "my-app",
									Env: []corev1.EnvVar{
										{Name: "DEPARTMENT",
											Value: "R&D",
										},
									},
								},
							},
						},
					},
					Namespace: "default",
				},
				ctx: context.Background(),
			},
			wantErr: false,
			wantPayload: corev1.Pod{
				ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "my-app",
							Env: []corev1.EnvVar{
								{
									Name:  LMAPMClusterName,
									Value: os.Getenv(ClusterName),
								},
								{
									Name:      LMAPMNodeName,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}},
								},
								{
									Name:      LMAPMPodName,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
								},
								{
									Name:      LMAPMPodNamespace,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
								},
								{
									Name:      LMAPMPodIP,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}},
								},
								{
									Name:      LMAPMPodUID,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.uid"}},
								},
								{
									Name:      ServiceNamespace,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels['app-namespace']"}},
								},
								{
									Name:      "SERVICE_ACCOUNT_NAME",
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.serviceAccountName"}},
								},
								{
									Name:      "SERVICE_NAME",
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels['app-name']"}},
								},
								{
									Name:  "DEPARTMENT",
									Value: "R&D",
								},
								{
									Name:  "OTLP_ENDPOINT",
									Value: "lmotel-svc:4317",
								},
								{
									Name:  "OTEL_RESOURCE_ATTRIBUTES",
									Value: "host.name=$(LM_APM_POD_NAME),ip=$(LM_APM_POD_IP),k8s.cluster.name=$(LM_APM_CLUSTER_NAME),k8s.namespace.name=$(LM_APM_POD_NAMESPACE),k8s.node.name=$(LM_APM_NODE_NAME),k8s.pod.uid=$(LM_APM_POD_UID),resource.type=kubernetes-pod,service.namespace=$(SERVICE_NAMESPACE),service.name=$(SERVICE_NAME),SERVICE_ACCOUNT_NAME=$(SERVICE_ACCOUNT_NAME)",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Mutate pod with external config with direct value of the SERVICE_NAMESPACE env var",
			args: struct {
				params *Params
				ctx    context.Context
			}{
				params: &Params{
					Client: k8sClient,
					Log:    logger,
					LMConfig: config.Config{
						MutationConfigProvided: true,
						MutationConfig: config.MutationConfig{
							LMEnvVars: config.LMEnvVars{
								Resource: []config.ResourceEnv{
									{
										Env: corev1.EnvVar{
											Name:  "SERVICE_NAMESPACE",
											Value: "us-west-2-techops",
										},
									},
									{
										Env: corev1.EnvVar{
											Name: "SERVICE_NAME",
											ValueFrom: &corev1.EnvVarSource{
												FieldRef: &corev1.ObjectFieldSelector{
													FieldPath: "metadata.labels['app-name']",
												},
											},
										},
									},
									{
										Env: corev1.EnvVar{
											Name: "SERVICE_ACCOUNT_NAME",
											ValueFrom: &corev1.EnvVarSource{
												FieldRef: &corev1.ObjectFieldSelector{
													FieldPath: "spec.serviceAccountName",
												},
											},
										},
									},
								},
								Operation: []config.OperationEnv{
									{
										Env: corev1.EnvVar{
											Name:  "OTLP_ENDPOINT",
											Value: "lmotel-svc:4317",
										},
									},
								},
							},
						},
					},
					Pod: &corev1.Pod{
						ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "my-app",
									Env: []corev1.EnvVar{
										{Name: "DEPARTMENT",
											Value: "R&D",
										},
									},
								},
							},
						},
					},
					Namespace: "default",
				},
				ctx: context.Background(),
			},
			wantErr: false,
			wantPayload: corev1.Pod{
				ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "my-app",
							Env: []corev1.EnvVar{
								{
									Name:  LMAPMClusterName,
									Value: os.Getenv(ClusterName),
								},
								{
									Name:      LMAPMNodeName,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}},
								},
								{
									Name:      LMAPMPodName,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
								},
								{
									Name:      LMAPMPodNamespace,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
								},
								{
									Name:      LMAPMPodIP,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}},
								},
								{
									Name:      LMAPMPodUID,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.uid"}},
								},
								{
									Name:  ServiceNamespace,
									Value: "us-west-2-techops",
								},
								{
									Name:      "SERVICE_NAME",
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels['app-name']"}},
								},
								{
									Name:      "SERVICE_ACCOUNT_NAME",
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.serviceAccountName"}},
								},
								{
									Name:  "DEPARTMENT",
									Value: "R&D",
								},
								{
									Name:  "OTLP_ENDPOINT",
									Value: "lmotel-svc:4317",
								},
								{
									Name:  "OTEL_RESOURCE_ATTRIBUTES",
									Value: "host.name=$(LM_APM_POD_NAME),ip=$(LM_APM_POD_IP),k8s.cluster.name=$(LM_APM_CLUSTER_NAME),k8s.namespace.name=$(LM_APM_POD_NAMESPACE),k8s.node.name=$(LM_APM_NODE_NAME),k8s.pod.uid=$(LM_APM_POD_UID),resource.type=kubernetes-pod,service.namespace=$(SERVICE_NAMESPACE),service.name=$(SERVICE_NAME),SERVICE_ACCOUNT_NAME=$(SERVICE_ACCOUNT_NAME)",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Mutate pod with external config with overriding of the value is allowed for some env var and disabled for some env var",
			args: struct {
				params *Params
				ctx    context.Context
			}{
				params: &Params{
					Client: k8sClient,
					Log:    logger,
					LMConfig: config.Config{
						MutationConfigProvided: true,
						MutationConfig: config.MutationConfig{
							LMEnvVars: config.LMEnvVars{
								Resource: []config.ResourceEnv{
									{
										Env: corev1.EnvVar{
											Name:  "SERVICE_NAMESPACE",
											Value: "us-west-2-techops",
										},
									},
									{
										Env: corev1.EnvVar{
											Name: "SERVICE_NAME",
											ValueFrom: &corev1.EnvVarSource{
												FieldRef: &corev1.ObjectFieldSelector{
													FieldPath: "metadata.labels['app-name']",
												},
											},
										},
									},
									{
										Env: corev1.EnvVar{
											Name: "SERVICE_ACCOUNT_NAME",
											ValueFrom: &corev1.EnvVarSource{
												FieldRef: &corev1.ObjectFieldSelector{
													FieldPath: "spec.serviceAccountName",
												},
											},
										},
									},
									{
										Env: corev1.EnvVar{
											Name:  "CLOUD_PROVIDER",
											Value: "AWS",
										},
										ResAttrName:      "cloud.provider",
										OverrideDisabled: true,
									},
								},
								Operation: []config.OperationEnv{
									{
										Env: corev1.EnvVar{
											Name:  "OTLP_ENDPOINT",
											Value: "lmotel-svc:4317",
										},
									},
									{
										Env: corev1.EnvVar{
											Name:  "COMPANY_NAME",
											Value: "ABC Corporation",
										},
										OverrideDisabled: true,
									},
								},
							},
						},
					},
					Pod: &corev1.Pod{
						ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "my-app",
									Env: []corev1.EnvVar{
										{
											Name:  "DEPARTMENT",
											Value: "R&D",
										},
										{
											Name:  "SERVICE_NAMESPACE",
											Value: "us-west-2-techops-prod",
										},
										{
											Name:  "SERVICE_NAME",
											Value: "test",
										},
										{
											Name:  "SERVICE_ACCOUNT_NAME",
											Value: "test",
										},
										{
											Name:  "OTLP_ENDPOINT",
											Value: "lmotel-svc:4318",
										},
									},
								},
							},
						},
					},
					Namespace: "default",
				},
				ctx: context.Background(),
			},
			wantErr: false,
			wantPayload: corev1.Pod{
				ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "my-app",
							Env: []corev1.EnvVar{
								{
									Name:  LMAPMClusterName,
									Value: os.Getenv(ClusterName),
								},
								{
									Name:      LMAPMNodeName,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}},
								},
								{
									Name:      LMAPMPodName,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
								},
								{
									Name:      LMAPMPodNamespace,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
								},
								{
									Name:      LMAPMPodIP,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}},
								},
								{
									Name:      LMAPMPodUID,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.uid"}},
								},
								{
									Name:  ServiceNamespace,
									Value: "us-west-2-techops-prod",
								},
								{
									Name:  ServiceName,
									Value: "test",
								},
								{
									Name:  "SERVICE_ACCOUNT_NAME",
									Value: "test",
								},
								{
									Name:  "CLOUD_PROVIDER",
									Value: "AWS",
								},
								{
									Name:  "OTLP_ENDPOINT",
									Value: "lmotel-svc:4318",
								},
								{
									Name:  "DEPARTMENT",
									Value: "R&D",
								},
								{
									Name:  "COMPANY_NAME",
									Value: "ABC Corporation",
								},
								{
									Name:  "OTEL_RESOURCE_ATTRIBUTES",
									Value: "host.name=$(LM_APM_POD_NAME),ip=$(LM_APM_POD_IP),k8s.cluster.name=$(LM_APM_CLUSTER_NAME),k8s.namespace.name=$(LM_APM_POD_NAMESPACE),k8s.node.name=$(LM_APM_NODE_NAME),k8s.pod.uid=$(LM_APM_POD_UID),resource.type=kubernetes-pod,service.namespace=$(SERVICE_NAMESPACE),service.name=$(SERVICE_NAME),SERVICE_ACCOUNT_NAME=$(SERVICE_ACCOUNT_NAME),cloud.provider=$(CLOUD_PROVIDER)",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Mutate pod with external config with label for SERVICE_NAME not found on pod",
			args: struct {
				params *Params
				ctx    context.Context
			}{
				params: &Params{
					Client: k8sClient,
					Log:    logger,
					LMConfig: config.Config{
						MutationConfigProvided: true,
						MutationConfig: config.MutationConfig{
							LMEnvVars: config.LMEnvVars{
								Resource: []config.ResourceEnv{
									{
										Env: corev1.EnvVar{
											Name: "SERVICE_NAMESPACE",
											ValueFrom: &corev1.EnvVarSource{
												FieldRef: &corev1.ObjectFieldSelector{
													FieldPath: "metadata.labels['app-namespace']",
												},
											},
										},
									},
									{
										Env: corev1.EnvVar{
											Name: "SERVICE_NAME",
											ValueFrom: &corev1.EnvVarSource{
												FieldRef: &corev1.ObjectFieldSelector{
													FieldPath: "metadata.labels['app-name']",
												},
											},
										},
									},
									{
										Env: corev1.EnvVar{
											Name: "SERVICE_ACCOUNT_NAME",
											ValueFrom: &corev1.EnvVarSource{
												FieldRef: &corev1.ObjectFieldSelector{
													FieldPath: "spec.serviceAccountName",
												},
											},
										},
									},
								},
								Operation: []config.OperationEnv{
									{
										Env: corev1.EnvVar{
											Name:  "OTLP_ENDPOINT",
											Value: "lmotel-svc:4317",
										},
									},
								},
							},
						},
					},
					Pod: &corev1.Pod{
						ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app": "test-app", "app-namespace": "test"}},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "my-app",
									Env: []corev1.EnvVar{
										{Name: "DEPARTMENT",
											Value: "R&D",
										},
									},
								},
							},
						},
					},
					Namespace: "default",
				},
				ctx: context.Background(),
			},
			wantErr: false,
			wantPayload: corev1.Pod{
				ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app": "test-app", "app-namespace": "test"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "my-app",
							Env: []corev1.EnvVar{
								{
									Name:  LMAPMClusterName,
									Value: os.Getenv(ClusterName),
								},
								{
									Name:      LMAPMNodeName,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}},
								},
								{
									Name:      LMAPMPodName,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
								},
								{
									Name:      LMAPMPodNamespace,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
								},
								{
									Name:      LMAPMPodIP,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}},
								},
								{
									Name:      LMAPMPodUID,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.uid"}},
								},
								{
									Name:      ServiceNamespace,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels['app-namespace']"}},
								},
								{
									Name:  "DEPARTMENT",
									Value: "R&D",
								},
								{
									Name:  ServiceName,
									Value: "test-pod",
								},
								{
									Name:      "SERVICE_ACCOUNT_NAME",
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.serviceAccountName"}},
								},
								{
									Name:  "OTEL_RESOURCE_ATTRIBUTES",
									Value: "host.name=$(LM_APM_POD_NAME),ip=$(LM_APM_POD_IP),k8s.cluster.name=$(LM_APM_CLUSTER_NAME),k8s.namespace.name=$(LM_APM_POD_NAMESPACE),k8s.node.name=$(LM_APM_NODE_NAME),k8s.pod.uid=$(LM_APM_POD_UID),resource.type=kubernetes-pod,service.namespace=$(SERVICE_NAMESPACE),service.name=$(SERVICE_NAME),SERVICE_ACCOUNT_NAME=$(SERVICE_ACCOUNT_NAME)",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Mutate pod with external config with LM reserved env variable passed from config",
			args: struct {
				params *Params
				ctx    context.Context
			}{
				params: &Params{
					Client: k8sClient,
					Log:    logger,
					LMConfig: config.Config{
						MutationConfigProvided: true,
						MutationConfig: config.MutationConfig{
							LMEnvVars: config.LMEnvVars{
								Resource: []config.ResourceEnv{
									{
										Env: corev1.EnvVar{
											Name: "SERVICE_NAMESPACE",
											ValueFrom: &corev1.EnvVarSource{
												FieldRef: &corev1.ObjectFieldSelector{
													FieldPath: "metadata.labels['app-namespace']",
												},
											},
										},
									},
									{
										Env: corev1.EnvVar{
											Name: "SERVICE_NAME",
											ValueFrom: &corev1.EnvVarSource{
												FieldRef: &corev1.ObjectFieldSelector{
													FieldPath: "metadata.labels['app-name']",
												},
											},
										},
									},
									{
										Env: corev1.EnvVar{
											Name: "SERVICE_ACCOUNT_NAME",
											ValueFrom: &corev1.EnvVarSource{
												FieldRef: &corev1.ObjectFieldSelector{
													FieldPath: "spec.serviceAccountName",
												},
											},
										},
									},
									{
										Env: corev1.EnvVar{
											Name:  "LM_APM_CLUSTER_NAME",
											Value: "default",
										},
									},
								},
								Operation: []config.OperationEnv{
									{
										Env: corev1.EnvVar{
											Name:  "OTLP_ENDPOINT",
											Value: "lmotel-svc:4317",
										},
									},
									{
										Env: corev1.EnvVar{
											Name:  "LM_APM_NODE_NAME",
											Value: "linux-node",
										},
									},
								},
							},
						},
					},
					Pod: &corev1.Pod{
						ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "my-app",
									Env: []corev1.EnvVar{
										{Name: "DEPARTMENT",
											Value: "R&D",
										},
									},
								},
							},
						},
					},
					Namespace: "default",
				},
			},
			wantErr: false,
			wantPayload: corev1.Pod{
				ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "my-app",
							Env: []corev1.EnvVar{
								{
									Name:  LMAPMClusterName,
									Value: os.Getenv(ClusterName),
								},
								{
									Name:      LMAPMNodeName,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}},
								},
								{
									Name:      LMAPMPodName,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
								},
								{
									Name:      LMAPMPodNamespace,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
								},
								{
									Name:      LMAPMPodIP,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}},
								},
								{
									Name:      LMAPMPodUID,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.uid"}},
								},
								{
									Name:      ServiceNamespace,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels['app-namespace']"}},
								},
								{
									Name:      "SERVICE_NAME",
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels['app-name']"}},
								},
								{
									Name:      "SERVICE_ACCOUNT_NAME",
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.serviceAccountName"}},
								},
								{
									Name:  "DEPARTMENT",
									Value: "R&D",
								},
								{
									Name:  "OTEL_RESOURCE_ATTRIBUTES",
									Value: "host.name=$(LM_APM_POD_NAME),ip=$(LM_APM_POD_IP),k8s.cluster.name=$(LM_APM_CLUSTER_NAME),k8s.namespace.name=$(LM_APM_POD_NAMESPACE),k8s.node.name=$(LM_APM_NODE_NAME),k8s.pod.uid=$(LM_APM_POD_UID),resource.type=kubernetes-pod,service.namespace=$(SERVICE_NAMESPACE),service.name=$(SERVICE_NAME),SERVICE_ACCOUNT_NAME=$(SERVICE_ACCOUNT_NAME)",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Mutate pod with external config with SERVICE_NAME or SERVICE_NAMESPACE passed in operations env var list",
			args: struct {
				params *Params
				ctx    context.Context
			}{
				params: &Params{
					Client: k8sClient,
					Log:    logger,
					LMConfig: config.Config{
						MutationConfigProvided: true,
						MutationConfig: config.MutationConfig{
							LMEnvVars: config.LMEnvVars{
								Resource: []config.ResourceEnv{
									{
										Env: corev1.EnvVar{
											Name: "SERVICE_ACCOUNT_NAME",
											ValueFrom: &corev1.EnvVarSource{
												FieldRef: &corev1.ObjectFieldSelector{
													FieldPath: "spec.serviceAccountName",
												},
											},
										},
									},
								},
								Operation: []config.OperationEnv{
									{
										Env: corev1.EnvVar{
											Name:  "OTLP_ENDPOINT",
											Value: "lmotel-svc:4317",
										},
									},
									{
										Env: corev1.EnvVar{
											Name: "SERVICE_NAME",
											ValueFrom: &corev1.EnvVarSource{
												FieldRef: &corev1.ObjectFieldSelector{
													FieldPath: "metadata.labels['app-name']",
												},
											},
										},
									},
									{
										Env: corev1.EnvVar{
											Name: "SERVICE_NAMESPACE",
											ValueFrom: &corev1.EnvVarSource{
												FieldRef: &corev1.ObjectFieldSelector{
													FieldPath: "metadata.labels['app-namespace']",
												},
											},
										},
									},
								},
							},
						},
					},
					Pod: &corev1.Pod{
						ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "my-app",
									Env: []corev1.EnvVar{
										{Name: "DEPARTMENT",
											Value: "R&D",
										},
									},
								},
							},
						},
					},
					Namespace: "default",
				},
				ctx: context.Background(),
			},
			wantErr: false,
			wantPayload: corev1.Pod{
				ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "my-app",
							Env: []corev1.EnvVar{
								{
									Name:  LMAPMClusterName,
									Value: os.Getenv(ClusterName),
								},
								{
									Name:      LMAPMNodeName,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}},
								},
								{
									Name:      LMAPMPodName,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
								},
								{
									Name:      LMAPMPodNamespace,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
								},
								{
									Name:      LMAPMPodIP,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}},
								},
								{
									Name:      LMAPMPodUID,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.uid"}},
								},
								{
									Name:      ServiceNamespace,
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
								},
								{
									Name:  ServiceName,
									Value: "test-pod",
								},
								{
									Name:      "SERVICE_ACCOUNT_NAME",
									ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.serviceAccountName"}},
								},
								{
									Name:  "DEPARTMENT",
									Value: "R&D",
								},
								{
									Name:  "OTEL_RESOURCE_ATTRIBUTES",
									Value: "host.name=$(LM_APM_POD_NAME),ip=$(LM_APM_POD_IP),k8s.cluster.name=$(LM_APM_CLUSTER_NAME),k8s.namespace.name=$(LM_APM_POD_NAMESPACE),k8s.node.name=$(LM_APM_NODE_NAME),k8s.pod.uid=$(LM_APM_POD_UID),resource.type=kubernetes-pod,service.namespace=$(SERVICE_NAMESPACE),SERVICE_ACCOUNT_NAME=$(SERVICE_ACCOUNT_NAME),service.name=$(SERVICE_NAME)",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mutateEnvVariables(context.Background(), tt.args.params)
			// t.Logf("%+v", tt.args.pod.Spec.Containers[0].Env)
			if (err != nil) != tt.wantErr {
				t.Errorf("MutatePod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, cont := range tt.args.params.Pod.Spec.Containers {
				for _, expectedEnv := range tt.wantPayload.Spec.Containers[0].Env {
					for _, env := range cont.Env {
						if expectedEnv.Name == env.Name {
							if expectedEnv.Value != "" {
								if !cmp.Equal(expectedEnv.Value, env.Value, cmpOpt) {
									t.Errorf("MutatePod() for environment variable %s, expected value is %s, but found %s", expectedEnv.Name, expectedEnv.Value, env.Value)
									return
								}
							}
							if expectedEnv.ValueFrom != nil {
								if !cmp.Equal(expectedEnv.ValueFrom, env.ValueFrom, cmpOpt) {
									t.Errorf("MutatePod() for environment variable %s, expected value is %s, but found %s", expectedEnv.Name, *expectedEnv.ValueFrom, *env.ValueFrom)
									return
								}
							}
						}
					}
				}
			}
		})
	}
}

func TestGetLmotelEnvironmentVariables(t *testing.T) {

	cmpOpt := cmp.AllowUnexported()

	os.Setenv("CLUSTER_NAME", "default")

	defer os.Unsetenv("CLUSTER_NAME")

	test := struct {
		name        string
		wantErr     bool
		wantPayload []corev1.EnvVar
	}{
		name:    "GetLmotelEnvironmentVariables",
		wantErr: false,
		wantPayload: []corev1.EnvVar{
			{
				Name:  LMAPMClusterName,
				Value: os.Getenv(ClusterName),
			},
			{
				Name:      LMAPMNodeName,
				ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}},
			},
			{
				Name:      LMAPMPodName,
				ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
			},
			{
				Name:      LMAPMPodNamespace,
				ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
			},
			{
				Name:      LMAPMPodIP,
				ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}},
			},
			{
				Name:      LMAPMPodUID,
				ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.uid"}},
			},
			{
				Name:      ServiceNamespace,
				ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
			},
			{
				Name:  "OTEL_RESOURCE_ATTRIBUTES",
				Value: "host.name=$(LM_APM_POD_NAME),ip=$(LM_APM_POD_IP),k8s.cluster.name=$(LM_APM_CLUSTER_NAME),k8s.namespace.name=$(LM_APM_POD_NAMESPACE),k8s.node.name=$(LM_APM_NODE_NAME),k8s.pod.uid=$(LM_APM_POD_UID),resource.type=kubernetes-pod,service.namespace=$(SERVICE_NAMESPACE)",
			},
		},
	}

	lmotelEnvVars := getLmotelEnvironmentVariables()

	if !cmp.Equal(lmotelEnvVars, test.wantPayload, cmpOpt) {
		t.Errorf("getLmotelEnvironmentVariables() expected value is %v, but found %v", test.wantPayload, lmotelEnvVars)
		return
	}
}

func TestMergeNewEnv(t *testing.T) {

	cmpOpt := cmp.AllowUnexported()

	os.Setenv("CLUSTER_NAME", "default")

	defer os.Unsetenv("CLUSTER_NAME")

	tests := []struct {
		name string
		args struct {
			originalEnvVars []corev1.EnvVar
			newEnvVars      []corev1.EnvVar
		}
		wantErr     bool
		wantPayload []corev1.EnvVar
	}{
		{
			name: "Merge new env variables with the original ones",
			args: struct {
				originalEnvVars []corev1.EnvVar
				newEnvVars      []corev1.EnvVar
			}{
				originalEnvVars: []corev1.EnvVar{
					{
						Name:  "DEPARTMENT",
						Value: "R&D",
					},
				},
				newEnvVars: []corev1.EnvVar{
					{
						Name:  LMAPMClusterName,
						Value: os.Getenv(ClusterName),
					},
					{
						Name:      LMAPMNodeName,
						ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}},
					},
					{
						Name:      LMAPMPodName,
						ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
					},
					{
						Name:      LMAPMPodNamespace,
						ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
					},
					{
						Name:      LMAPMPodIP,
						ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}},
					},
					{
						Name:      LMAPMPodUID,
						ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.uid"}},
					},
					{
						Name:      ServiceNamespace,
						ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
					},
					{
						Name:  ServiceName,
						Value: "test-pod",
					},
					{
						Name:  "OTEL_RESOURCE_ATTRIBUTES",
						Value: "resource.type=kubernetes-pod,ip=$(LM_APM_POD_IP),host.name=$(LM_APM_POD_NAME),k8s.pod.uid=$(LM_APM_POD_UID),service.namespace=$(SERVICE_NAMESPACE),k8s.namespace.name=$(LM_APM_POD_NAMESPACE),k8s.node.name=$(LM_APM_NODE_NAME),k8s.cluster.name=$(LM_APM_CLUSTER_NAME),service.name=$(SERVICE_NAME)",
					},
				},
			},
			wantErr: false,
			wantPayload: []corev1.EnvVar{
				{
					Name:  "DEPARTMENT",
					Value: "R&D",
				},
				{
					Name:  LMAPMClusterName,
					Value: os.Getenv(ClusterName),
				},
				{
					Name:      LMAPMNodeName,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}},
				},
				{
					Name:      LMAPMPodName,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
				},
				{
					Name:      LMAPMPodNamespace,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
				},
				{
					Name:      LMAPMPodIP,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}},
				},
				{
					Name:      LMAPMPodUID,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.uid"}},
				},
				{
					Name:      ServiceNamespace,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
				},
				{
					Name:  ServiceName,
					Value: "test-pod",
				},
				{
					Name:  "OTEL_RESOURCE_ATTRIBUTES",
					Value: "resource.type=kubernetes-pod,ip=$(LM_APM_POD_IP),host.name=$(LM_APM_POD_NAME),k8s.pod.uid=$(LM_APM_POD_UID),service.namespace=$(SERVICE_NAMESPACE),k8s.namespace.name=$(LM_APM_POD_NAMESPACE),k8s.node.name=$(LM_APM_NODE_NAME),k8s.cluster.name=$(LM_APM_CLUSTER_NAME),service.name=$(SERVICE_NAME)",
				},
			},
		},
		{
			name: "Merge new env variables conflicting with original env vars",
			args: struct {
				originalEnvVars []corev1.EnvVar
				newEnvVars      []corev1.EnvVar
			}{
				originalEnvVars: []corev1.EnvVar{
					{
						Name:  "DEPARTMENT",
						Value: "R&D",
					},
					{
						Name:      ServiceNamespace,
						ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels['app-namespace']"}},
					},
					{
						Name:  OTELResourceAttributes,
						Value: "k8s.container.name=nginx",
					},
				},
				newEnvVars: []corev1.EnvVar{
					{
						Name:  LMAPMClusterName,
						Value: os.Getenv(ClusterName),
					},
					{
						Name:      LMAPMNodeName,
						ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}},
					},
					{
						Name:      LMAPMPodName,
						ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
					},
					{
						Name:      LMAPMPodNamespace,
						ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
					},
					{
						Name:      LMAPMPodIP,
						ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}},
					},
					{
						Name:      LMAPMPodUID,
						ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.uid"}},
					},
					{
						Name:      ServiceNamespace,
						ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
					},
					{
						Name:  ServiceName,
						Value: "test-pod",
					},
					{
						Name:  "OTEL_RESOURCE_ATTRIBUTES",
						Value: "resource.type=kubernetes-pod,ip=$(LM_APM_POD_IP),host.name=$(LM_APM_POD_NAME),k8s.pod.uid=$(LM_APM_POD_UID),service.namespace=$(SERVICE_NAMESPACE),k8s.namespace.name=$(LM_APM_POD_NAMESPACE),k8s.node.name=$(LM_APM_NODE_NAME),k8s.cluster.name=$(LM_APM_CLUSTER_NAME),service.name=$(SERVICE_NAME)",
					},
				},
			},
			wantErr: false,
			wantPayload: []corev1.EnvVar{
				{
					Name:  "DEPARTMENT",
					Value: "R&D",
				},
				{
					Name:  LMAPMClusterName,
					Value: os.Getenv(ClusterName),
				},
				{
					Name:      LMAPMNodeName,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}},
				},
				{
					Name:      LMAPMPodName,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
				},
				{
					Name:      LMAPMPodNamespace,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
				},
				{
					Name:      LMAPMPodIP,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}},
				},
				{
					Name:      LMAPMPodUID,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.uid"}},
				},
				{
					Name:      ServiceNamespace,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
				},
				{
					Name:  ServiceName,
					Value: "test-pod",
				},
				{
					Name:  "OTEL_RESOURCE_ATTRIBUTES",
					Value: "resource.type=kubernetes-pod,ip=$(LM_APM_POD_IP),host.name=$(LM_APM_POD_NAME),k8s.pod.uid=$(LM_APM_POD_UID),service.namespace=$(SERVICE_NAMESPACE),k8s.namespace.name=$(LM_APM_POD_NAMESPACE),k8s.node.name=$(LM_APM_NODE_NAME),k8s.cluster.name=$(LM_APM_CLUSTER_NAME),service.name=$(SERVICE_NAME),k8s.container.name=nginx",
				},
			},
		},
	}

	for _, tt := range tests {
		mergedEnvVars, err := mergeNewEnv(tt.args.originalEnvVars, tt.args.newEnvVars)

		if (err != nil) != tt.wantErr {
			t.Errorf("mergeNewEnv() return error = %v, but expected error = %v", err, tt.wantErr)
			return
		}

		if len(mergedEnvVars) != len(tt.wantPayload) {
			t.Errorf("mergeNewEnv() returned %d number of env variables, but expected env number of env variables = %d", len(mergedEnvVars), len(tt.wantPayload))
			return
		}

		for _, expectedEnvVar := range tt.wantPayload {
			for _, mergedEnvVar := range mergedEnvVars {
				if expectedEnvVar.Name == mergedEnvVar.Name {
					if expectedEnvVar.Value != "" {
						if !cmp.Equal(expectedEnvVar.Value, mergedEnvVar.Value, cmpOpt) {
							t.Errorf("MutatePod() for environment variable %s, expected value is %s, but found %s", expectedEnvVar.Name, expectedEnvVar.Value, mergedEnvVar.Value)
							return
						}
					}
					if expectedEnvVar.ValueFrom != nil {
						if !cmp.Equal(expectedEnvVar.ValueFrom, mergedEnvVar.ValueFrom, cmpOpt) {
							t.Errorf("MutatePod() for environment variable %s, expected value is %s, but found %s", expectedEnvVar.Name, *expectedEnvVar.ValueFrom, *mergedEnvVar.ValueFrom)
							return
						}
					}
				}
			}
		}
	}
}

func TestAddResEnvToOtelResAttribute(t *testing.T) {

	cmpOpt := cmp.AllowUnexported()

	os.Setenv("CLUSTER_NAME", "default")

	defer os.Unsetenv("CLUSTER_NAME")

	test := struct {
		name string
		args struct {
			resourceEnvVar corev1.EnvVar
			newEnvVars     []corev1.EnvVar
		}
		wantPayload struct {
			envVars []corev1.EnvVar
		}
	}{
		name: "Add resource env variable to the OTEL_RESOURCE_ATTRIBUTE environment variable",
		args: struct {
			resourceEnvVar corev1.EnvVar
			newEnvVars     []corev1.EnvVar
		}{
			resourceEnvVar: corev1.EnvVar{
				Name:      "SERVICE_NAME",
				ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels['app-namespace']"}},
			},
			newEnvVars: []corev1.EnvVar{
				{
					Name:  LMAPMClusterName,
					Value: os.Getenv(ClusterName),
				},
				{
					Name:      LMAPMNodeName,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}},
				},
				{
					Name:      LMAPMPodName,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
				},
				{
					Name:      LMAPMPodNamespace,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
				},
				{
					Name:      LMAPMPodIP,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}},
				},
				{
					Name:      LMAPMPodUID,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.uid"}},
				},
				{
					Name:      ServiceNamespace,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels['app-namespace']"}},
				},
				{
					Name:  "OTEL_RESOURCE_ATTRIBUTES",
					Value: "resource.type=kubernetes-pod,ip=$(LM_APM_POD_IP),host.name=$(LM_APM_POD_NAME),k8s.pod.uid=$(LM_APM_POD_UID),service.namespace=$(SERVICE_NAMESPACE),k8s.namespace.name=$(LM_APM_POD_NAMESPACE),k8s.node.name=$(LM_APM_NODE_NAME),k8s.cluster.name=$(LM_APM_CLUSTER_NAME)",
				},
			},
		},
		wantPayload: struct {
			envVars []corev1.EnvVar
		}{

			envVars: []corev1.EnvVar{
				{
					Name:  LMAPMClusterName,
					Value: os.Getenv(ClusterName),
				},
				{
					Name:      LMAPMNodeName,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}},
				},
				{
					Name:      LMAPMPodName,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"}},
				},
				{
					Name:      LMAPMPodNamespace,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
				},
				{
					Name:      LMAPMPodIP,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "status.podIP"}},
				},
				{
					Name:      LMAPMPodUID,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.uid"}},
				},
				{
					Name:      ServiceNamespace,
					ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels['app-namespace']"}},
				},
				{
					Name:  "OTEL_RESOURCE_ATTRIBUTES",
					Value: "resource.type=kubernetes-pod,ip=$(LM_APM_POD_IP),host.name=$(LM_APM_POD_NAME),k8s.pod.uid=$(LM_APM_POD_UID),service.namespace=$(SERVICE_NAMESPACE),k8s.namespace.name=$(LM_APM_POD_NAMESPACE),k8s.node.name=$(LM_APM_NODE_NAME),k8s.cluster.name=$(LM_APM_CLUSTER_NAME),service.name=$(SERVICE_NAME)",
				},
			},
		},
	}

	newEnvVars := addResEnvToOtelResAttribute(test.args.resourceEnvVar, test.args.newEnvVars, "")

	for _, expectedEnvVar := range test.wantPayload.envVars {
		for _, envVar := range newEnvVars {
			if expectedEnvVar.Name == envVar.Name {

				if expectedEnvVar.Value != "" {
					if !cmp.Equal(expectedEnvVar.Value, envVar.Value, cmpOpt) {
						t.Errorf("AddResEnvToOtelResAttribute() for environment variable %s, expected value is %s, but found %s", expectedEnvVar.Name, expectedEnvVar.Value, envVar.Value)
						return
					}
				}
				if expectedEnvVar.ValueFrom != nil {
					if !cmp.Equal(expectedEnvVar.ValueFrom, envVar.ValueFrom, cmpOpt) {
						t.Errorf("AddResEnvToOtelResAttribute() for environment variable %s, expected value is %s, but found %s", expectedEnvVar.Name, *expectedEnvVar.ValueFrom, *envVar.ValueFrom)
						return
					}
				}
			}
		}
	}
}

func TestGetParentWorkloadNameForPod(t *testing.T) {
	k8sClient, err := getFakeK8sClient()
	if err != nil {
		t.Errorf("Error occurred in getting fake k8s client: %v", err)
		return
	}

	os.Setenv("CLUSTER_NAME", "default")
	defer os.Unsetenv("CLUSTER_NAME")

	tests := []struct {
		name string
		args struct {
			pod       *corev1.Pod
			k8sClient *config.K8sClient
			namespace string
		}
		wantErr     bool
		wantPayload string
	}{
		{
			name: "Get parent workload name for bare pod",
			args: struct {
				pod       *corev1.Pod
				k8sClient *config.K8sClient
				namespace string
			}{
				&corev1.Pod{
					ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}, OwnerReferences: nil},
					Spec:       corev1.PodSpec{},
				},
				k8sClient,
				"default",
			},
			wantErr:     false,
			wantPayload: "test-pod",
		},
		{
			name: "Get parent workload name for k8s Job",
			args: struct {
				pod       *corev1.Pod
				k8sClient *config.K8sClient
				namespace string
			}{
				pod: &corev1.Pod{
					ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}, OwnerReferences: []v1.OwnerReference{{Name: "hello-job", Kind: "Job"}}},
					Spec:       corev1.PodSpec{},
				},
				k8sClient: k8sClient,
				namespace: "default",
			},
			wantErr:     false,
			wantPayload: "hello-job",
		},
		{
			name: "Get parent workload name for k8s DaemonSet",
			args: struct {
				pod       *corev1.Pod
				k8sClient *config.K8sClient
				namespace string
			}{
				pod: &corev1.Pod{
					ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}, OwnerReferences: []v1.OwnerReference{{Name: "hello-daemonSet", Kind: "DaemonSet"}}},
					Spec:       corev1.PodSpec{},
				},
				k8sClient: k8sClient,
				namespace: "default",
			},
			wantErr:     false,
			wantPayload: "hello-daemonSet",
		},
		{
			name: "Get parent workload name for k8s StatefulSet",
			args: struct {
				pod       *corev1.Pod
				k8sClient *config.K8sClient
				namespace string
			}{
				pod: &corev1.Pod{
					ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}, OwnerReferences: []v1.OwnerReference{{Name: "hello-statefulSet", Kind: "StatefulSet"}}},
					Spec:       corev1.PodSpec{},
				},
				k8sClient: k8sClient,
				namespace: "default",
			},
			wantErr:     false,
			wantPayload: "hello-statefulSet",
		},
		{
			name: "Get parent workload name for k8s ReplicaSet",
			args: struct {
				pod       *corev1.Pod
				k8sClient *config.K8sClient
				namespace string
			}{
				pod: &corev1.Pod{
					ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}, OwnerReferences: []v1.OwnerReference{{Name: "hello-replicaSet", Kind: "ReplicaSet"}}},
					Spec:       corev1.PodSpec{},
				},
				k8sClient: k8sClient,
				namespace: "default",
			},
			wantErr:     false,
			wantPayload: "hello-replicaSet",
		},
		{
			name: "Get parent workload name for incorrect owner reference",
			args: struct {
				pod       *corev1.Pod
				k8sClient *config.K8sClient
				namespace string
			}{
				pod: &corev1.Pod{
					ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}, OwnerReferences: []v1.OwnerReference{{Name: "hello-1", Kind: "ReplicaSet"}}},
					Spec:       corev1.PodSpec{},
				},
				k8sClient: k8sClient,
				namespace: "default",
			},
			wantErr:     true,
			wantPayload: "",
		},
		{
			name: "Get parent workload name for incorrect owner kind",
			args: struct {
				pod       *corev1.Pod
				k8sClient *config.K8sClient
				namespace string
			}{
				pod: &corev1.Pod{
					ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}, OwnerReferences: []v1.OwnerReference{{Name: "hello", Kind: "NotExistKind"}}},
					Spec:       corev1.PodSpec{},
				},
				k8sClient: k8sClient,
				namespace: "default",
			},
			wantErr:     true,
			wantPayload: "",
		},
	}

	for _, tt := range tests {
		workloadName, err := getParentWorkloadNameForPod(tt.args.pod, tt.args.k8sClient, tt.args.namespace)

		if (err != nil) != tt.wantErr {
			t.Errorf("getParentWorkloadNameForPod() error = %v, but expected is error = %v", err, tt.wantErr)
			return
		}

		if workloadName != tt.wantPayload {
			t.Errorf("getParentWorkloadNameForPod() returned workloadName = %v, but expected is workloadName = %v", workloadName, tt.wantPayload)
			return
		}
	}
}

func TestExtractResourceWorkloadName(t *testing.T) {
	k8sClient, err := getFakeK8sClient()
	if err != nil {
		t.Errorf("Error occurred in getting fake k8s client: %v", err)
		return
	}

	os.Setenv("CLUSTER_NAME", "default")
	defer os.Unsetenv("CLUSTER_NAME")

	tests := []struct {
		name string
		args struct {
			owner          client.Object
			k8sClient      *config.K8sClient
			namespacedName types.NamespacedName
		}
		wantErr     bool
		wantPayload string
	}{
		{
			name: "Extract resource workload name for replicaSets",
			args: struct {
				owner          client.Object
				k8sClient      *config.K8sClient
				namespacedName types.NamespacedName
			}{
				owner:          &appsv1.ReplicaSet{},
				k8sClient:      k8sClient,
				namespacedName: types.NamespacedName{Namespace: "default", Name: "hello-replicaSet"},
			},
			wantErr:     false,
			wantPayload: "hello-replicaSet",
		},
		{
			name: "Extract resource workload name for statefulSets",
			args: struct {
				owner          client.Object
				k8sClient      *config.K8sClient
				namespacedName types.NamespacedName
			}{
				owner:          &appsv1.StatefulSet{},
				k8sClient:      k8sClient,
				namespacedName: types.NamespacedName{Namespace: "default", Name: "hello-statefulSet"},
			},
			wantErr:     false,
			wantPayload: "hello-statefulSet",
		},
		{
			name: "Extract resource workload name for DaemonSets",
			args: struct {
				owner          client.Object
				k8sClient      *config.K8sClient
				namespacedName types.NamespacedName
			}{
				owner:          &appsv1.DaemonSet{},
				k8sClient:      k8sClient,
				namespacedName: types.NamespacedName{Namespace: "default", Name: "hello-daemonSet"},
			},
			wantErr:     false,
			wantPayload: "hello-daemonSet",
		},
		{
			name: "Extract resource workload name for Jobs",
			args: struct {
				owner          client.Object
				k8sClient      *config.K8sClient
				namespacedName types.NamespacedName
			}{
				owner:          &batchv1.Job{},
				k8sClient:      k8sClient,
				namespacedName: types.NamespacedName{Namespace: "default", Name: "hello-job"},
			},
			wantErr:     false,
			wantPayload: "hello-job",
		},
		{
			name: "Extract resource workload name for replicaSets managed by deployment",
			args: struct {
				owner          client.Object
				k8sClient      *config.K8sClient
				namespacedName types.NamespacedName
			}{
				owner:          &appsv1.ReplicaSet{},
				k8sClient:      k8sClient,
				namespacedName: types.NamespacedName{Namespace: "default", Name: "hello-replicaSetManagedByDeployment"},
			},
			wantErr:     false,
			wantPayload: "hello-deployment",
		},
		{
			name: "Extract resource workload name for invalid owner name",
			args: struct {
				owner          client.Object
				k8sClient      *config.K8sClient
				namespacedName types.NamespacedName
			}{
				owner:          &appsv1.ReplicaSet{},
				k8sClient:      k8sClient,
				namespacedName: types.NamespacedName{Namespace: "default", Name: "something-owner"},
			},
			wantErr:     true,
			wantPayload: "",
		},
	}

	for _, tt := range tests {
		workloadName, err := extractResourceWorkloadName(tt.args.namespacedName, tt.args.k8sClient, tt.args.owner)

		if (err != nil) != tt.wantErr {
			t.Errorf("extractResourceWorkloadName() error = %v, but expected is error = %v", err, tt.wantErr)
			return
		}

		if workloadName != tt.wantPayload {
			t.Errorf("extractResourceWorkloadName() returned workloadName = %v, but expected is workloadName = %v", workloadName, tt.wantPayload)
			return
		}
	}
}

func TestCheckIfPodHasLabel(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			pod    *corev1.Pod
			envVar corev1.EnvVar
		}
		wantErr     bool
		wantPayload struct {
			lableValue string
			found      bool
			err        error
		}
	}{
		{
			name: "Check if pod has label specified in env variable",
			args: struct {
				pod    *corev1.Pod
				envVar corev1.EnvVar
			}{
				pod: &corev1.Pod{
					ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}},
					Spec:       corev1.PodSpec{},
				},
				envVar: corev1.EnvVar{
					Name: "SERVICE_NAME",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "metadata.labels['app-name']",
						},
					},
				},
			},
			wantErr: false,
			wantPayload: struct {
				lableValue string
				found      bool
				err        error
			}{
				lableValue: "test-app",
				found:      true,
				err:        nil,
			},
		},
		{
			name: "If pod does not have label specified in env variable",
			args: struct {
				pod    *corev1.Pod
				envVar corev1.EnvVar
			}{
				pod: &corev1.Pod{
					ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}},
					Spec:       corev1.PodSpec{},
				},
				envVar: corev1.EnvVar{
					Name: "SERVICE_NAME",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "metadata.labels['application-name']",
						},
					},
				},
			},
			wantErr: true,
			wantPayload: struct {
				lableValue string
				found      bool
				err        error
			}{
				lableValue: "",
				found:      false,
				err:        errEnvVarValueLabelNotFoundOnPod,
			},
		},
		{
			name: "If env variable value is not specified in label path format",
			args: struct {
				pod    *corev1.Pod
				envVar corev1.EnvVar
			}{
				pod: &corev1.Pod{
					ObjectMeta: v1.ObjectMeta{Name: "test-pod", Labels: map[string]string{"app-name": "test-app", "app-namespace": "test"}},
					Spec:       corev1.PodSpec{},
				},
				envVar: corev1.EnvVar{
					Name: "SERVICE_NAME",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "test-config-map"}, Key: "SERVICE_NAME"},
					},
				},
			},
			wantErr: true,
			wantPayload: struct {
				lableValue string
				found      bool
				err        error
			}{
				lableValue: "",
				found:      false,
				err:        errEnvVarValueNotInLabelBasedFieldPathFormat,
			},
		},
	}

	for _, tt := range tests {
		labelValue, _, err := checkIfPodHasLabel(tt.args.pod, tt.args.envVar)

		if (err != nil) != tt.wantErr {
			t.Errorf("checkIfPodHasLabel() error = %v, but expected is error = %v", err, tt.wantErr)
			return
		}

		if err != nil && err != tt.wantPayload.err {
			t.Errorf("checkIfPodHasLabel() error = %v, but expected is error = %v", err, tt.wantPayload.err)
			return
		}

		if labelValue != tt.wantPayload.lableValue {
			t.Errorf("labelValue() returned labelName = %v, but expected is labelName = %v", labelValue, tt.wantPayload.lableValue)
			return
		}
	}
}

func TestGetOTELSemVarKey(t *testing.T) {

	tests := []struct {
		name string
		args struct {
			rawKey string
		}
		wantPayload struct {
			otelSemVarKey string
			found         bool
		}
	}{
		{
			name: "get OTEL sem var key for SERVICE_NAMESPACE",
			args: struct{ rawKey string }{
				"SERVICE_NAMESPACE",
			},
			wantPayload: struct {
				otelSemVarKey string
				found         bool
			}{
				"service.namespace",
				true,
			},
		},
		{
			name: "get OTEL sem var key for SERVICE_NAME",
			args: struct{ rawKey string }{
				"SERVICE_NAME",
			},
			wantPayload: struct {
				otelSemVarKey string
				found         bool
			}{
				"service.name",
				true,
			},
		},
		{
			name: "get OTEL sem var key for UNKNOWN (Not found)",
			args: struct{ rawKey string }{
				"UNKNOWN",
			},
			wantPayload: struct {
				otelSemVarKey string
				found         bool
			}{
				"",
				false,
			},
		},
	}

	for _, tt := range tests {
		otelKey, found := getOTELSemVarKey(tt.args.rawKey)

		if otelKey != tt.wantPayload.otelSemVarKey || found != tt.wantPayload.found {
			t.Errorf("getOTELSemVarKey() returns otelSemVarKey = %v & found = %v, but expected is otelSemVarKey = %v & found = %v", otelKey, found, tt.wantPayload.otelSemVarKey, tt.wantPayload.found)
			return
		}
	}
}

func TestRunMutations(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			ctx    context.Context
			params *Params
		}
		wantErr bool
	}{
		{
			name: "Test RunMutations",
			args: struct {
				ctx    context.Context
				params *Params
			}{
				ctx: context.Background(),
				params: &Params{
					Client:    &config.K8sClient{},
					Pod:       &corev1.Pod{ObjectMeta: v1.ObjectMeta{Name: "demo"}, Spec: corev1.PodSpec{Containers: []corev1.Container{{Image: "nginx", Name: "nginx"}}}},
					LMConfig:  config.Config{},
					Mutations: Mutations,
					Namespace: "default",
					Log:       logger,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {

		err := RunMutations(tt.args.ctx, tt.args.params)

		if (err != nil) != tt.wantErr {
			t.Errorf("RunMutations() error = %v, but expected is error = %v", err, tt.wantErr)
			return
		}
	}
}

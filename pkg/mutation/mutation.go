package mutation

import (
	"context"
	"errors"

	"github.com/logicmonitor/lm-k8s-webhook/pkg/config"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
)

const (
	// Resource attributes
	LMAPMClusterName       = "LM_APM_CLUSTER_NAME"
	LMAPMNodeName          = "LM_APM_NODE_NAME"
	LMAPMPodName           = "LM_APM_POD_NAME"
	LMAPMPodNamespace      = "LM_APM_POD_NAMESPACE"
	LMAPMPodIP             = "LM_APM_POD_IP"
	LMAPMPodUID            = "LM_APM_POD_UID"
	ClusterName            = "CLUSTER_NAME"
	ServiceNamespace       = "SERVICE_NAMESPACE"
	ServiceName            = "SERVICE_NAME"
	OTELResourceAttributes = "OTEL_RESOURCE_ATTRIBUTES"

	// Workload resource discovery
	WorkloadResourceDeployment  = "Deployment"
	WorkloadResourceStatefulSet = "StatefulSet"
	WorkloadResourceDaemonSet   = "DaemonSet"
	WorkloadResourceReplicaSet  = "ReplicaSet"
	WorkloadResourceJob         = "Job"
	WorkloadResourceCronJob     = "CronJob"

	// Mutation
	MutationEnvVarInjection = "envVarInjection"
)

// var ignoredNamespaces = []string{
// 	metav1.NamespaceSystem,
// 	metav1.NamespacePublic,
// }

// skipList represents the env variables that the user should not pass through external config or manifest, these are managed by webhook itself
var skipList = []string{LMAPMClusterName, LMAPMNodeName, LMAPMPodName, LMAPMPodNamespace, LMAPMPodIP, LMAPMPodUID, OTELResourceAttributes}

// errors
var (
	errEnvVarValueNotInLabelBasedFieldPathFormat = errors.New("environment variable value is not specified in label based field path format")
	errEnvVarValueLabelNotFoundOnPod             = errors.New("label specified in environment variable is not found on pod or value pointed by label is empty")
)

type Mutation struct {
	Name string
	Do   func(context.Context, *Params) error
}

var Mutations = []Mutation{{Name: MutationEnvVarInjection, Do: mutateEnvVariables}}

type Params struct {
	Client    *config.K8sClient
	Log       logr.Logger
	LMConfig  config.Config
	Mutations []Mutation
	Pod       *corev1.Pod
	Namespace string
}

func RunMutations(ctx context.Context, params *Params) error {
	for _, mutation := range params.Mutations {
		if mutationRequired(mutation, params.Pod) {
			err := mutation.Do(ctx, params)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func mutationRequired(mutation Mutation, pod *corev1.Pod) bool {
	// TODO:Need to add checks if mutation is to be done or not
	return true
}

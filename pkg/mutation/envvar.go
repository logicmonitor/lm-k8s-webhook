package mutation

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/logicmonitor/lm-k8s-webhook/pkg/config"
	"go.opentelemetry.io/otel/semconv"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func mutateEnvVariables(ctx context.Context, params *Params) error {

	var isServiceNameEnvFound bool
	var otelResourceAttributesIndex int

	logger := log.Log.WithValues("mutate-pod", fmt.Sprintf("%s/%s", params.Namespace, params.Pod.GetName()))

	newEnvVars := getLmotelEnvironmentVariables()

	// If external config is provided then only perform this operation
	if params.LMConfig.MutationConfigProvided {
		logger.Info("As external config present, checking for new env vars")
		var isEnvVarToBeSkipped bool
		var isServiceNamespaceEnvFound bool

		otelResourceAttributesIndex = len(newEnvVars) - 1

		for _, resourceEnvVar := range params.LMConfig.MutationConfig.LMEnvVars.Resource {
			isEnvVarToBeSkipped = false
			isServiceNamespaceEnvFound = false
			// Check if resourceEnvVar is a part of skipList, if present in skip list then skip that env variable
			for _, skipListEnvvar := range skipList {
				if skipListEnvvar == resourceEnvVar.Name {
					isEnvVarToBeSkipped = true
					logger.Info("Skipped resource env variable", "env var", resourceEnvVar.Name, "env value", resourceEnvVar.Value, "env valueFrom", resourceEnvVar.ValueFrom)
					break
				}
			}

			// If env variable is not in skip list
			// add as a new env variable to the env list
			if !isEnvVarToBeSkipped {

				// If resourceEnvVar is SERVICE_NAMESPACE
				if resourceEnvVar.Name == ServiceNamespace {
					isServiceNamespaceEnvFound = true

					if resourceEnvVar.Value != "" {
						// Direct value is passed
						for i, envVar := range newEnvVars {
							if envVar.Name == ServiceNamespace {
								newEnvVars[i] = resourceEnvVar
								logger.Info("resourceEnvVar is ServiceNamespace, overriding the default value of ServiceNamespace", "env value", resourceEnvVar.Value)
								break
							}
						}
					} else if resourceEnvVar.ValueFrom != nil {
						_, found, err := checkIfPodHasLabel(params.Pod, resourceEnvVar)

						// Update SERVICE_NAMESPACE env var either if the label is present on pod or value is not specified in metadata.label format
						if found || (err == errEnvVarValueNotInLabelBasedFieldPathFormat) {
							for i, envVar := range newEnvVars {
								if envVar.Name == ServiceNamespace {
									newEnvVars[i] = resourceEnvVar
									logger.Info("resourceEnvVar is ServiceNamespace, overriding the default value of ServiceNamespace", "env valueFrom", resourceEnvVar.ValueFrom)
									break
								}
							}
						}
					}
				}

				if isServiceNamespaceEnvFound {
					continue
				}

				// If resourceEnvVar is SERVICE_NAME
				if resourceEnvVar.Name == ServiceName {
					isServiceNameEnvFound = true
					if resourceEnvVar.ValueFrom != nil {
						podLabelValue, found, _ := checkIfPodHasLabel(params.Pod, resourceEnvVar)

						// Update SERVICE_NAME env var either if the label is present on pod or value is not specified in metadata.label format
						if !found || (len(strings.Trim(podLabelValue, " "))) == 0 {
							logger.Info("deriving the SERVICE_NAME value from workload resource")
							workloadResource, _ := getParentWorkloadNameForPod(params.Pod, params.Client, params.Namespace)
							svcNameEnv := corev1.EnvVar{Name: resourceEnvVar.Name, Value: workloadResource}
							newEnvVars = append(newEnvVars, svcNameEnv)
							// Add it to the OTELResourceAttributes
							newEnvVars, otelResourceAttributesIndex = addResEnvToOtelResAttribute(svcNameEnv, newEnvVars)
							continue
						}
					}
				}

				logger.Info("Adding new resource env variable", "Name: ", resourceEnvVar.Name, "env value", resourceEnvVar.Value, "env valueFrom", resourceEnvVar.ValueFrom)
				newEnvVars = append(newEnvVars, resourceEnvVar)

				// Add it to the OTELResourceAttributes
				newEnvVars, otelResourceAttributesIndex = addResEnvToOtelResAttribute(resourceEnvVar, newEnvVars)
			}
		}

		for _, operationEnvVar := range params.LMConfig.MutationConfig.LMEnvVars.Operation {
			isEnvVarToBeSkipped = false
			// Check if operationEnvVar is a part of skipList, if present in skip list then skip that env variable
			for _, skipListEnvvar := range skipList {
				if skipListEnvvar == operationEnvVar.Name {
					isEnvVarToBeSkipped = true
					logger.Info("Skipped operation env variable", "Name:", operationEnvVar.Name)
					break
				}
			}

			// If operationEnvVar is SERVICE_NAMESPACE
			if operationEnvVar.Name == ServiceNamespace {
				logger.Info("operationEnvVar is ServiceNamespace, skipping it as ServiceNamespace should be the part of resource environment variables")
				isEnvVarToBeSkipped = true
			}

			// If operationEnvVar is SERVICE_NAME
			if operationEnvVar.Name == ServiceName {
				logger.Info("operationEnvVar is ServiceName, skipping it as ServiceName should be the part of resource environment variables")
				isEnvVarToBeSkipped = true
			}

			// If env variable is not in skip list
			// add as a new env variable to the env list
			if !isEnvVarToBeSkipped {
				logger.Info("Added new operation env variable", "Name:", operationEnvVar.Name, "env.value", operationEnvVar.Value, "env.ValueFrom", operationEnvVar.ValueFrom)
				newEnvVars = append(newEnvVars, operationEnvVar)
			}
		}
	}

	// If SERVICE_NAME env is not found then add it
	if !isServiceNameEnvFound {
		workloadResource, _ := getParentWorkloadNameForPod(params.Pod, params.Client, params.Namespace)
		svcNameEnv := corev1.EnvVar{Name: ServiceName, Value: workloadResource}
		newEnvVars = append(newEnvVars, svcNameEnv)
		// Add it to the OTELResourceAttributes
		newEnvVars, otelResourceAttributesIndex = addResEnvToOtelResAttribute(svcNameEnv, newEnvVars)
	}

	// Check if OTEL_RESOURCE_ATTRIBUTES is not the last element already, then no need to move it
	if otelResourceAttributesIndex != len(newEnvVars)-1 {

		// Move the OTELResourceAttributes to the last index of the newEnvVar list to satisfy environment variable dependency

		// Copy last env var in the list
		currentLastEnvVar := newEnvVars[len(newEnvVars)-1]

		// Move OTELResourceAttributes env variable to the last
		newEnvVars[len(newEnvVars)-1] = newEnvVars[otelResourceAttributesIndex]

		// Move previously copied last element to the OTELResourceAttributes's old position
		newEnvVars[otelResourceAttributesIndex] = currentLastEnvVar
	}

	for i, ctr := range params.Pod.Spec.Containers {
		// TODO: Mutate only the specific container
		envVars, err := mergeNewEnv(ctr.Env, newEnvVars)
		if err != nil {
			return err
		}
		ctr.Env = envVars
		params.Pod.Spec.Containers[i] = ctr
		logger.Info("Final list of env variables after merge", "env vars:", envVars)
	}
	return nil
}

// getLmotelEnvironmentVariables returns a list of default env variables required by LM-OTEL
func getLmotelEnvironmentVariables() []corev1.EnvVar {

	// Creates a list of default env variables required by LM-OTEL
	lmotelEnvVars := []corev1.EnvVar{
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
		// For now we are passing the pod namespace value to the service namespace
		{
			Name:      ServiceNamespace,
			ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
		},
	}

	var buffer strings.Builder

	// OTELResourceAttributes are constructed with string builder for string concatnation memory optimization

	buffer.WriteString("resource.type=kubernetes-pod,")
	buffer.WriteString("ip=$(")
	buffer.WriteString(LMAPMPodIP)
	buffer.WriteString("),host.name=$(")
	buffer.WriteString(LMAPMPodName)
	buffer.WriteString("),k8s.pod.uid=$(")
	buffer.WriteString(LMAPMPodUID)
	buffer.WriteString("),service.namespace=$(")
	buffer.WriteString(ServiceNamespace)
	buffer.WriteString("),k8s.namespace.name=$(")
	buffer.WriteString(LMAPMPodNamespace)
	buffer.WriteString("),k8s.node.name=$(")
	buffer.WriteString(LMAPMNodeName)
	buffer.WriteString("),k8s.cluster.name=$(")
	buffer.WriteString(LMAPMClusterName)
	buffer.WriteString(")")

	lmotelEnvVars = append(lmotelEnvVars, corev1.EnvVar{Name: OTELResourceAttributes, Value: buffer.String()})

	return lmotelEnvVars
}

// Merges new environment variables with the existing ones
func mergeNewEnv(originalEnvVars []corev1.EnvVar, newEnvVars []corev1.EnvVar) ([]corev1.EnvVar, error) {
	logger := log.Log.WithName("mergeNewEnv")

	origEnvVarMap := map[string]corev1.EnvVar{}
	for _, v := range originalEnvVars {
		origEnvVarMap[v.Name] = v
	}
	mergedEnv := make([]corev1.EnvVar, len(originalEnvVars))
	copy(mergedEnv, originalEnvVars)

	// Check if new env var is already there in the pod definition
	for _, newEnvVar := range newEnvVars {
		envVar, ok := origEnvVarMap[newEnvVar.Name]
		if !ok {
			// if we dont have already, append it
			origEnvVarMap[newEnvVar.Name] = newEnvVar
			mergedEnv = append(mergedEnv, newEnvVar)
			continue
		}

		if !reflect.DeepEqual(envVar, newEnvVar) {
			logger.Info("Property conflict found", newEnvVar.Name, newEnvVar.Value, newEnvVar.Name, newEnvVar.ValueFrom)

			// Check if SERVICE_NAMESPACE is set already
			// if set then assign the env var set by user in manifest
			if envVar.Name == ServiceNamespace {
				logger.Info("ServiceNamespace is overriden from the manifest", envVar.Name, envVar.Value, envVar.Name, envVar.ValueFrom)
				for i, envVar := range newEnvVars {
					if envVar.Name == ServiceNamespace {
						newEnvVars[i] = envVar
						break
					}
				}
			}
		}
	}

	return mergedEnv, nil
}

// getOTELSemVarKey returns the key as per the OTEL semantic conventions for the given raw key
func getOTELSemVarKey(rawKey string) (string, bool) {
	switch rawKey {
	case ServiceNamespace:
		return string(semconv.ServiceNamespaceKey), true

	case ServiceName:
		return string(semconv.ServiceNameKey), true

	default:
		return "", false
	}
}

// getParentWorkloadNameForPod returns the parent workload name which is managing the pod
func getParentWorkloadNameForPod(pod *corev1.Pod, k8sClient *config.K8sClient, namespace string) (string, error) {
	logger := log.Log.WithName("getParentWorkloadNameForPod")
	// If no owner reference is present, that means Pod is deployed independently
	if len(pod.GetOwnerReferences()) == 0 {
		logger.Info("Orphan pod is found")
		return pod.GetObjectMeta().GetName(), nil
	}

	ownerRef := pod.GetOwnerReferences()[0]

	namespacedName := types.NamespacedName{Namespace: namespace, Name: ownerRef.Name}
	switch ownerRef.Kind {
	case WorkloadResourceJob:
		var owner batchv1.Job
		return extractResourceWorkloadName(namespacedName, k8sClient, &owner)
	case WorkloadResourceReplicaSet:
		var owner appsv1.ReplicaSet
		return extractResourceWorkloadName(namespacedName, k8sClient, &owner)
	case WorkloadResourceDaemonSet:
		var owner appsv1.DaemonSet
		return extractResourceWorkloadName(namespacedName, k8sClient, &owner)
	case WorkloadResourceStatefulSet:
		var owner appsv1.StatefulSet
		return extractResourceWorkloadName(namespacedName, k8sClient, &owner)
	}
	return "", fmt.Errorf("invalid workload resource: %s", ownerRef.Kind)
}

// extractResourceWorkloadName extracts the resource workload name of the pod based on the owner references
func extractResourceWorkloadName(namespacedName types.NamespacedName, k8sClient *config.K8sClient, owner client.Object) (string, error) {
	logger := log.Log.WithName("extractResourceWorkloadName")
	getOpts := metav1.GetOptions{}

	var err error

	switch owner.(type) {
	case *appsv1.ReplicaSet:
		owner, err = k8sClient.Clientset.AppsV1().ReplicaSets(namespacedName.Namespace).Get(context.Background(), namespacedName.Name, getOpts)
	case *appsv1.DaemonSet:
		owner, err = k8sClient.Clientset.AppsV1().DaemonSets(namespacedName.Namespace).Get(context.Background(), namespacedName.Name, getOpts)
	case *appsv1.StatefulSet:
		owner, err = k8sClient.Clientset.AppsV1().StatefulSets(namespacedName.Namespace).Get(context.Background(), namespacedName.Name, getOpts)
	case *batchv1.Job:
		owner, err = k8sClient.Clientset.BatchV1().Jobs(namespacedName.Namespace).Get(context.Background(), namespacedName.Name, getOpts)
	default:
		return "", fmt.Errorf("invalid workload resource type: %s", owner.GetObjectKind())
	}
	if err != nil {
		logger.Error(err, "error in getting owner resource details")
		return "", err
	}
	if owner.GetOwnerReferences() != nil && len(owner.GetOwnerReferences()) > 0 && owner.GetOwnerReferences()[0].Name != "" {
		logger.Info("Workload Name", "Workload name", owner.GetOwnerReferences()[0].Name)
		return owner.GetOwnerReferences()[0].Name, nil
	}
	return namespacedName.Name, nil
}

// addResEnvToOtelResAttribute adds resource env variable to the OTELResourceAttributes
func addResEnvToOtelResAttribute(resourceEnvVar corev1.EnvVar, newEnvVars []corev1.EnvVar) ([]corev1.EnvVar, int) {
	otelResourceAttributesIndex := 0
	for i, envVar := range newEnvVars {
		// Find the location of OTELResourceAttributes in the list
		if envVar.Name == OTELResourceAttributes {
			otelResourceAttributesIndex = i

			var buffer strings.Builder

			buffer.WriteString(envVar.Value)
			buffer.WriteString(",")
			serviceNameOTELSemVarKey, found := getOTELSemVarKey(resourceEnvVar.Name)
			if found {
				buffer.WriteString(serviceNameOTELSemVarKey)
			} else {
				buffer.WriteString(resourceEnvVar.Name)
			}
			buffer.WriteString("=$(")
			buffer.WriteString(resourceEnvVar.Name)
			buffer.WriteString(")")

			// Update the OTELResourceAttributes value with the updated one
			newEnvVars[otelResourceAttributesIndex].Value = buffer.String()
			break
		}
	}
	return newEnvVars, otelResourceAttributesIndex
}

// checkIfPodHasLabel checks if label specified in a env value is present on the pod
func checkIfPodHasLabel(pod *corev1.Pod, envVar corev1.EnvVar) (string, bool, error) {
	logger := log.Log.WithName("checkIfPodHasLabel")
	// Parse label name
	exp, err := regexp.Compile(`\[\'(.*?)\'\]`)
	if err != nil {
		logger.Error(err, "Invalid regex")
		return "", false, err
	} else {
		if envVar.ValueFrom.FieldRef != nil {
			matchedStrings := exp.FindStringSubmatch(envVar.ValueFrom.FieldRef.FieldPath)
			if len(matchedStrings) > 1 {
				podLabelValue, found := pod.Labels[matchedStrings[1]]

				// If specified label is not present on pod or if its value is empty
				if !found || (len(strings.Trim(podLabelValue, " ")) == 0) {
					logger.Error(err, "cannot find the label-name specified in "+envVar.Name+" environment variable value metadata.labels['label-name'] on pod.")
					return "", false, errEnvVarValueLabelNotFoundOnPod
				}
				return podLabelValue, true, nil
			}
		}
		return "", false, errEnvVarValueNotInLabelBasedFieldPathFormat
	}
}

package mutation

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sort"
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

	var isServiceNameEnvProcessed bool
	var isServiceNamespaceEnvProcessed bool
	var otelResourceAttributesIndex int

	logger := log.Log.WithValues("mutate-pod", fmt.Sprintf("%s/%s", params.Namespace, params.Pod.GetName()))

	newEnvVars := getLmotelEnvironmentVariables()

	// Get application container
	container := getApplicationContainer(params.Pod)

	// If external config is provided then only perform this operation
	if params.LMConfig.MutationConfigProvided {
		logger.Info("As external config present, checking for new env vars")
		var isEnvVarToBeSkipped bool

		otelResourceAttributesIndex = len(newEnvVars) - 1

		for _, resourceEnvVar := range params.LMConfig.MutationConfig.LMEnvVars.Resource {
			isEnvVarToBeSkipped = false
			// isServiceNamespaceEnvProcessed = false
			// Check if resourceEnvVar is a part of skipList, if present in skip list then skip that env variable
			for _, skipListEnvvar := range skipList {
				if skipListEnvvar == resourceEnvVar.Env.Name {
					isEnvVarToBeSkipped = true
					logger.Info("Skipped resource env variable", "env var", resourceEnvVar.Env.Name, "env value", resourceEnvVar.Env.Name, "env valueFrom", resourceEnvVar.Env.ValueFrom)
					break
				}
			}

			// If env variable is not in skip list
			// add as a new env variable to the env list
			if !isEnvVarToBeSkipped {

				// If resourceEnvVar is SERVICE_NAMESPACE
				if resourceEnvVar.Env.Name == ServiceNamespace {
					// isServiceNamespaceEnvFound = true
					// If override is allowed
					if !resourceEnvVar.OverrideDisabled {
						if idx := getIndexOfEnv(container.Env, ServiceNamespace); idx > -1 {
							svcNamespaceIdx := getIndexOfEnv(newEnvVars, ServiceNamespace)
							svcNamespaceEnv := corev1.EnvVar{Name: resourceEnvVar.Env.Name, Value: container.Env[idx].Value, ValueFrom: container.Env[idx].ValueFrom}
							newEnvVars[svcNamespaceIdx] = svcNamespaceEnv
							isServiceNamespaceEnvProcessed = true
							logger.Info("resourceEnvVar is SERVICE_NAMESPACE, overriding the default value of SERVICE_NAMESPACE from container", "env value", newEnvVars[svcNamespaceIdx].Value)
							continue
						}
					}

					if resourceEnvVar.Env.Value != "" {
						// Direct value is passed
						svcNamespaceIdx := getIndexOfEnv(newEnvVars, ServiceNamespace)
						newEnvVars[svcNamespaceIdx] = resourceEnvVar.Env
						isServiceNamespaceEnvProcessed = true
						logger.Info("resourceEnvVar is SERVICE_NAMESPACE, overriding the default value of SERVICE_NAMESPACE", "env value", resourceEnvVar.Env.Value)
						continue
					}

					if resourceEnvVar.Env.ValueFrom != nil {
						_, found, err := checkIfPodHasLabel(params.Pod, resourceEnvVar.Env)

						// Update SERVICE_NAMESPACE env var either if the label is present on pod or value is not specified in metadata.label format
						if found || (err == errEnvVarValueNotInLabelBasedFieldPathFormat) {
							svcNamespaceIdx := getIndexOfEnv(newEnvVars, ServiceNamespace)
							newEnvVars[svcNamespaceIdx] = resourceEnvVar.Env
							isServiceNamespaceEnvProcessed = true
							logger.Info("resourceEnvVar is SERVICE_NAMESPACE, overriding the default value of ServiceNamespace", "env valueFrom", resourceEnvVar.Env.ValueFrom)
						}
						continue
					}
				}

				// If resourceEnvVar is SERVICE_NAME
				if resourceEnvVar.Env.Name == ServiceName {
					// If override is allowed
					if !resourceEnvVar.OverrideDisabled {
						if idx := getIndexOfEnv(container.Env, ServiceName); idx > -1 {
							svcNameEnv := corev1.EnvVar{Name: resourceEnvVar.Env.Name, Value: container.Env[idx].Value, ValueFrom: container.Env[idx].ValueFrom}
							newEnvVars = append(newEnvVars, svcNameEnv)

							// Add it to the OTELResourceAttributes
							newEnvVars, otelResourceAttributesIndex = addResEnvToOtelResAttribute(svcNameEnv, newEnvVars, resourceEnvVar.ResAttrName)
							isServiceNameEnvProcessed = true
							logger.Info("resourceEnvVar is SERVICE_NAME, using value of the SERVICE_NAME from container", "SERVICE_NAME env:", svcNameEnv)
							continue
						}
					}

					if resourceEnvVar.Env.ValueFrom != nil {
						podLabelValue, found, err := checkIfPodHasLabel(params.Pod, resourceEnvVar.Env)

						// Update SERVICE_NAME env var either if the label is present on pod or value is not specified in metadata.label format
						if found || (err == errEnvVarValueNotInLabelBasedFieldPathFormat) {
							newEnvVars = append(newEnvVars, resourceEnvVar.Env)

							// Add it to the OTELResourceAttributes
							newEnvVars, otelResourceAttributesIndex = addResEnvToOtelResAttribute(resourceEnvVar.Env, newEnvVars, resourceEnvVar.ResAttrName)
							isServiceNameEnvProcessed = true
							logger.Info("resourceEnvVar is SERVICE_NAME", "SERVICE_NAME env:", resourceEnvVar.Env)
							continue
						}

						if !found || (len(strings.Trim(podLabelValue, " "))) == 0 {
							logger.Info("deriving the SERVICE_NAME value from workload resource")
							workloadResource, _ := getParentWorkloadNameForPod(params.Pod, params.Client, params.Namespace)
							svcNameEnv := corev1.EnvVar{Name: resourceEnvVar.Env.Name, Value: workloadResource}
							newEnvVars = append(newEnvVars, svcNameEnv)

							// Add it to the OTELResourceAttributes
							newEnvVars, otelResourceAttributesIndex = addResEnvToOtelResAttribute(svcNameEnv, newEnvVars, resourceEnvVar.ResAttrName)
							isServiceNameEnvProcessed = true
							logger.Info("resourceEnvVar is SERVICE_NAME, using value of the SERVICE_NAME from workload resource", "SERVICE_NAME env:", svcNameEnv)
							continue
						}
					}
				}

				// For any other env var
				var envToBeAdded corev1.EnvVar
				if !resourceEnvVar.OverrideDisabled {
					// if the env is present in application container already, then use it
					if idx := getIndexOfEnv(container.Env, resourceEnvVar.Env.Name); idx > -1 {
						envToBeAdded = container.Env[idx]
					} else {
						envToBeAdded = resourceEnvVar.Env
					}
				} else {
					envToBeAdded = resourceEnvVar.Env
				}
				newEnvVars = append(newEnvVars, envToBeAdded)

				// Add it to the OTELResourceAttributes
				newEnvVars, otelResourceAttributesIndex = addResEnvToOtelResAttribute(envToBeAdded, newEnvVars, resourceEnvVar.ResAttrName)
				logger.Info("Adding new resource env variable", "Name: ", envToBeAdded.Name, "env value", envToBeAdded.Value, "env valueFrom", envToBeAdded.ValueFrom)
			}
		}

		for _, operationEnvVar := range params.LMConfig.MutationConfig.LMEnvVars.Operation {
			isEnvVarToBeSkipped = false
			// Check if operationEnvVar is a part of skipList, if present in skip list then skip that env variable
			for _, skipListEnvvar := range skipList {
				if skipListEnvvar == operationEnvVar.Env.Name {
					isEnvVarToBeSkipped = true
					logger.Info("Skipped operation env variable", "Name:", operationEnvVar.Env.Name)
					break
				}
			}

			// If operationEnvVar is SERVICE_NAMESPACE
			if operationEnvVar.Env.Name == ServiceNamespace {
				logger.Info("operationEnvVar is SERVICE_NAMESPACE, skipping it as ServiceNamespace should be the part of resource environment variables")
				isEnvVarToBeSkipped = true
			}

			// If operationEnvVar is SERVICE_NAME
			if operationEnvVar.Env.Name == ServiceName {
				logger.Info("operationEnvVar is SERVICE_NAME, skipping it as ServiceName should be the part of resource environment variables")
				isEnvVarToBeSkipped = true
			}

			// If env variable is not in skip list
			// add as a new env variable to the env list
			if !isEnvVarToBeSkipped {
				// for any other env var
				var envToBeAdded corev1.EnvVar
				if !operationEnvVar.OverrideDisabled {
					// if the env is present in application container already, then use it
					if idx := getIndexOfEnv(container.Env, operationEnvVar.Env.Name); idx > -1 {
						envToBeAdded = container.Env[idx]
					} else {
						envToBeAdded = operationEnvVar.Env
					}
				} else {
					envToBeAdded = operationEnvVar.Env
				}
				newEnvVars = append(newEnvVars, envToBeAdded)
				logger.Info("Added new operation env variable", "Name:", envToBeAdded.Name, "env.value", envToBeAdded.Value, "env.ValueFrom", envToBeAdded.ValueFrom)
			}
		}
	}

	if !isServiceNamespaceEnvProcessed {
		if idx := getIndexOfEnv(container.Env, ServiceNamespace); idx > -1 {
			svcNamespaceIdx := getIndexOfEnv(newEnvVars, ServiceNamespace)
			svcNamespaceEnv := corev1.EnvVar{Name: ServiceNamespace, Value: container.Env[idx].Value, ValueFrom: container.Env[idx].ValueFrom}
			newEnvVars[svcNamespaceIdx] = svcNamespaceEnv
			logger.Info("resourceEnvVar is SERVICE_NAMESPACE, using value from container", "env value", svcNamespaceEnv)
		}
	}

	// If SERVICE_NAME env is not found then add it
	if !isServiceNameEnvProcessed {
		// Check if present in the container
		// If present, then use that value otherwise derive from workload

		if idx := getIndexOfEnv(container.Env, ServiceName); idx > -1 {
			svcNameEnv := corev1.EnvVar{Name: ServiceName, Value: container.Env[idx].Value, ValueFrom: container.Env[idx].ValueFrom}
			newEnvVars = append(newEnvVars, svcNameEnv)
			// Add it to the OTELResourceAttributes
			newEnvVars, otelResourceAttributesIndex = addResEnvToOtelResAttribute(svcNameEnv, newEnvVars, "")
			logger.Info("resourceEnvVar is SERVICE_NAME, using value from container", "env value", svcNameEnv)
		} else {
			workloadResource, _ := getParentWorkloadNameForPod(params.Pod, params.Client, params.Namespace)
			svcNameEnv := corev1.EnvVar{Name: ServiceName, Value: workloadResource}
			newEnvVars = append(newEnvVars, svcNameEnv)
			// Add it to the OTELResourceAttributes
			newEnvVars, otelResourceAttributesIndex = addResEnvToOtelResAttribute(svcNameEnv, newEnvVars, "")
			logger.Info("resourceEnvVar is SERVICE_NAME, derived value from workload", "env value", svcNameEnv)
		}
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

	envVars, err := mergeNewEnv(container.Env, newEnvVars)
	if err != nil {
		return err
	}
	logger.Info("Final list of env variables after merge", "env vars:", envVars)
	for idx, ctr := range params.Pod.Spec.Containers {
		if ctr.Name == container.Name {
			ctr.Env = envVars
			params.Pod.Spec.Containers[idx] = ctr
			break
		}
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

	res := map[string]string{}

	res["resource.type"] = "kubernetes-pod"
	res["ip"] = fmt.Sprintf("$(%s)", LMAPMPodIP)
	res["host.name"] = fmt.Sprintf("$(%s)", LMAPMPodName)
	res["k8s.pod.uid"] = fmt.Sprintf("$(%s)", LMAPMPodUID)
	res["service.namespace"] = fmt.Sprintf("$(%s)", ServiceNamespace)
	res["k8s.namespace.name"] = fmt.Sprintf("$(%s)", LMAPMPodNamespace)
	res["k8s.node.name"] = fmt.Sprintf("$(%s)", LMAPMNodeName)
	res["k8s.cluster.name"] = fmt.Sprintf("$(%s)", LMAPMClusterName)

	resStr := createResMapStr(res)

	lmotelEnvVars = append(lmotelEnvVars, corev1.EnvVar{Name: OTELResourceAttributes, Value: resStr})

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
			logger.Info("env var conflict found", "newEnvVar", newEnvVar, "envVar", envVar)

			idx := getIndexOfEnv(mergedEnv, envVar.Name)
			mergedEnv[idx].Value = newEnvVar.Value
			mergedEnv[idx].ValueFrom = newEnvVar.ValueFrom

			if envVar.Name == OTELResourceAttributes {
				// get OTEL_RESOURCE_ATTRIBUTES from new env var list
				newOTELResAttrIndex := getIndexOfEnv(newEnvVars, OTELResourceAttributes)
				existingNewOtelResAttr := strings.Split(newEnvVars[newOTELResAttrIndex].Value, ",")
				newOtelResAttr := make(map[string]bool)

				for _, attr := range existingNewOtelResAttr {
					attrKeyValue := strings.Split(attr, "=")
					newOtelResAttr[attrKeyValue[0]] = true
				}

				existingCtrOtelResAttr := strings.Split(envVar.Value, ",")

				for _, attr := range existingCtrOtelResAttr {
					attrKeyValue := strings.Split(attr, "=")
					// If resource attribute from OTEL_RESOURCE_ATTRIBUTES in container is not found in new OTEL_RESOURCE_ATTRIBUTES then add it
					if !newOtelResAttr[attrKeyValue[0]] {
						newEnvVars[newOTELResAttrIndex].Value = newEnvVars[newOTELResAttrIndex].Value + "," + attr
					}
				}

				ctrOtelResAttrIndex := getIndexOfEnv(mergedEnv, OTELResourceAttributes)
				// Update the ctr env var OTEL_RESOURCE_ATTRIBUTES with the updated value
				mergedEnv[ctrOtelResAttrIndex].Value = newEnvVars[newOTELResAttrIndex].Value
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

	for _, ownerRef := range pod.GetOwnerReferences() {
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
	}
	return "", fmt.Errorf("invalid workload resource: %v", pod.GetOwnerReferences())
}

// extractResourceWorkloadName extracts the resource workload name of the pod based on the owner references
func extractResourceWorkloadName(namespacedName types.NamespacedName, k8sClient *config.K8sClient, owner client.Object) (string, error) {
	logger := log.Log.WithName("extractResourceWorkloadName")
	getOpts := metav1.GetOptions{}

	var err error

	switch owner.(type) {
	case *appsv1.ReplicaSet:
		owner, err = k8sClient.Clientset.AppsV1().ReplicaSets(namespacedName.Namespace).Get(context.Background(), namespacedName.Name, getOpts)
		if err != nil {
			logger.Error(err, "error in getting owner resource details")
			return "", err
		}
		for _, parentOwnerRef := range owner.GetOwnerReferences() {
			if parentOwnerRef.Kind == WorkloadResourceDeployment {
				var parentOwner appsv1.Deployment
				return extractResourceWorkloadName(types.NamespacedName{Namespace: owner.GetNamespace(), Name: parentOwnerRef.Name}, k8sClient, &parentOwner)
			}
		}
	case *appsv1.Deployment:
		return namespacedName.Name, nil

	case *appsv1.DaemonSet:
		return namespacedName.Name, nil

	case *appsv1.StatefulSet:
		return namespacedName.Name, nil

	case *batchv1.Job:
		return namespacedName.Name, nil
	}
	return namespacedName.Name, nil
}

// addResEnvToOtelResAttribute adds resource env variable to the OTELResourceAttributes
func addResEnvToOtelResAttribute(resourceEnvVar corev1.EnvVar, newEnvVars []corev1.EnvVar, resAttrName string) ([]corev1.EnvVar, int) {
	var otelResourceAttributesIndex int
	var newEnvStr string
	// Find the location of OTELResourceAttributes in the list
	otelResourceAttributesIndex = getIndexOfEnv(newEnvVars, OTELResourceAttributes)
	if otelResourceAttributesIndex > -1 {
		otelSemVarKey, found := getOTELSemVarKey(resourceEnvVar.Name)
		if found {
			newEnvStr = fmt.Sprintf("%s=$(%s)", otelSemVarKey, resourceEnvVar.Name)
		} else {
			if resAttrName != "" {
				newEnvStr = fmt.Sprintf("%s=$(%s)", resAttrName, resourceEnvVar.Name)
			} else {
				newEnvStr = fmt.Sprintf("%s=$(%s)", resourceEnvVar.Name, resourceEnvVar.Name)
			}
		}
		// Update the OTELResourceAttributes value with the updated one
		newEnvVars[otelResourceAttributesIndex].Value = fmt.Sprintf("%s,%s", newEnvVars[otelResourceAttributesIndex].Value, newEnvStr)
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

func createResMapStr(res map[string]string) string {
	resKeys := make([]string, 0, len(res))
	for key := range res {
		resKeys = append(resKeys, key)
	}
	sort.Strings(resKeys)
	var resString string
	for _, reskey := range resKeys {
		if resString != "" {
			resString += ","
		}
		resString += fmt.Sprintf("%s=%s", reskey, res[reskey])
	}
	return resString
}

func getIndexOfEnv(envs []corev1.EnvVar, name string) int {
	for i := range envs {
		if envs[i].Name == name {
			return i
		}
	}
	return -1
}

func getApplicationContainer(pod *corev1.Pod) corev1.Container {
	return pod.Spec.Containers[0]
}

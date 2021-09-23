
# LM-Webhook

-----

**LM-Webhook** is the implementation of the Kubenetes Mutating Admission webhook. Some of the key features of the LM-Webhook are:

- LM-Webhook can be used to inject the kubernetes specific resource attributes like pod name, ip, pod namespace, service namespace, pod UUID in the pod as an environment variables, which avoids the need of manually updating the deployment manifests to include these resource attributes. 
- Custom environment variables can also be injected by passing the external configuration.    

-----
## Setup:

helm-chart which is provided in this repo, installs the lm-webhook in the Kubernetes Cluster.  

## Prerequisites:

* Kubernetes cluster
* Helm is required for the deployment of lm-webhook witn helm-chart  

In Kubernetes, in order for the API server to communicate with the webhook component, the webhook requires a TLS certificate that the API server is configured to trust.

- #### TLS certificate management using cert-manager:
    You can refer to [cert-manager installation](https://cert-manager.io/docs/installation/) for TLS the certificate setup using cert-manager

-----
## Using Selectors

Selectors can be used to limit which requests can be intercepted by the webhook based on the labels.
Two types of selectors can be specified in _MutatingWebhookConfiguration_  i.e. _ObjectSelector_ and _NamespaceSelector_.

Both ObjectSelector and NamespaceSelector can use matchLabels and matchExpressions to specify the selectors.
You can check [working with kubernetes objects and labels](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) for more details 

#### 1. ObjectSelector:

ObjectSelector is used to specify the label based selectors for the objects (pod) for which the requests are required to be intercepted.

**Example:** 
Using matchLabels, objectSelector can be specified as follows:

```
 objectSelector:
    matchLabels:
      tier: backend
```

With this selectors the requests for objects (pod) with label tier = backend will be intercepted.      
Using matchExpressions, objectSelector can be specified as follows:

```
objectSelector:
  matchExpressions:
    - key: tier
      operator: In
      values: ["frontend","backend"]
```

With this selectors the requests for objects (pod) with label tier = backend or tier = frontend will be intercepted.
#### 2. NamespaceSelector:

NamespaceSelector is used to specify the label based selectors for the namespaces.

**Example:** 
Using matchLabels, namespaceSelector can be specified as follows:

```
 namespaceSelector:
    matchLabels:
      environment: development
```
With this selectors the requests for objects (pods) with label tier = backend will be intercepted.      
Using matchExpressions, namespaceSelector can be specified as follows:

```
namespaceSelector:
  matchExpressions:
    - key: tier
      operator: In
      values: ["development","staging"]
```
-----
## Deploying the lm-webhook helm-chart

You can refer the following demo command for deploying the lm-webhook with the helm-chart.
Run following command in the bash terminal from the helm-chart/lm-webhook directory.

```
helm install --debug --wait -n lm-webhook \
--create-namespace \
--set cluster_name="default" \
--set mutatingWebhook.objectSelector.matchLabels.tier="backend" \
--set mutatingWebhook.objectSelector.matchExpressions[0].key="type" \
--set mutatingWebhook.objectSelector.matchExpressions[0].operator="In" \
--set mutatingWebhook.objectSelector.matchExpressions[0].values[0]=application \
--set mutatingWebhook.objectSelector.matchExpressions[0].values[1]=service \
--set mutatingWebhook.namespaceSelector.matchExpressions[0].key="environment" \
--set mutatingWebhook.namespaceSelector.matchExpressions[0].operator="In" \
--set mutatingWebhook.namespaceSelector.matchExpressions[0].values[0]="dev" \
--set-file lmconfig=<path_to_external_config> \
lm-webhook .
```

#### Required Values:

- **cluster_name (default: ""):** Name of the k8s cluster in which lm-webhook will be deployed.

#### Optional Values:

- **mutatingWebhook.objectSelector (default: ""):** specifies the label based selectors for the objects (pod) for which the requests are required to be intercepted
- **mutatingWebhook.namespaceSelector (default: ""):** specifies the label based selectors for the namespaces.
- **lmconfig (default: ""):** specifies the external config file path

Selectors used in the above example command states that, only the object (pod) creation requests with the object (pod) having lables as tier="backend" and type=["application" or "service"] 
which belong to the namespaces having label as environment=["dev" or "staging"] will be intercepted by the lm-webhook.

-----
#### External config

Currently as a part of the external config, user can define the custom environment variables that are to be injected into the application pods.

**Example:**
```
lmEnvVars:
  resource:
    - name: SERVICE_ACCOUNT_NAME
      valueFrom:
        fieldRef:
          fieldPath: spec.serviceAccountName
    - name: SERVICE_NAMESPACE
      valueFrom:
        fieldRef:
          fieldPath: metadata.labels['app-namespace']
    - name: SERVICE_NAME
      valueFrom:
        fieldRef:
          fieldPath: metadata.labels['app-name']
  operation:
    - name: COMPANY_NAME
      value: ABC Corporation
    - name: OTLP_ENDPOINT
      value: lmotel-svc:55680
    - name: OTEL_JAVAAGENT_ENABLED
      value: true
```

environment variables can be of two types, i.e. resource and operation
- Resource holds the resource environment variables, which will be the part of _OTEL_RESOURCE_ATTRIBUTES_.
- Operation holds the operation environment variables, which will not be the part of _OTEL_RESOURCE_ATTRIBUTES_ but can be used in the application for custom use cases.

LM-Webhook injects following environment variables in the application pods. It is not recommanded to explicitely specify these environment variables as a part of pod definition. Only SERVICE_NAMESPACE can be overriden, either by specifying it as a part of pod definition or in the external configuration. Default value of SERVICE_NAMESPACE is the value of the pod namespace. 

Values for SERVICE_NAME and SERVICE_NAMESPACE can also be specified in terms of pod label as shown in above example config. So that value of the specified pod label can be used as a SERVICE_NAME or SERVICE_NAMESPACE. 


| SR. No. | Environment Variable Name | 
| ---: | :--- | 
| 1 | LM_APM_CLUSTER_NAME | 
| 2 | LM_APM_NODE_NAME |
| 3 | LM_APM_POD_NAME |
| 4 | LM_APM_POD_IP | 
| 5 | LM_APM_POD_NAMESPACE | 
| 6 | LM_APM_POD_UID | 
| 7 | SERVICE_NAMESPACE | 
| 8 | OTEL_RESOURCE_ATTRIBUTES | 

-----

### License

 This Source Code is subject to the terms of the Mozilla Public
  License, v. 2.0. If a copy of the MPL was not distributed with this
  file, You can obtain one at https://mozilla.org/MPL/2.0/.
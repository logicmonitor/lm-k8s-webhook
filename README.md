
# LM-Webhook

-----

**LM-Webhook** is the implementation of the Kubenetes Mutating Admission webhook. Some of the key features of the LM-Webhook are:

- LM-Webhook can be used to inject the kubernetes specific resource attributes like pod name, ip, pod namespace, service namespace, pod UUID in the pod as an environment variables, which avoids the need of manually updating the deployment manifests to include these resource attributes. 
- Custom environment variables can also be injected by passing the external configuration.    

-----
## Setup:

helm-chart which is provided in this repo, installs the lm-webhook in the Kubernetes Cluster.  

## Prerequisites:

* Kubernetes: Please refer [Kubernetes admission webhook prerequisites](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#prerequisites)
* Helm 3.0+ is required for the deployment of lm-webhook with helm-chart  

### TLS Certificate Requirement:

In Kubernetes, in order for the API server to communicate with the webhook component, the webhook requires a TLS certificate that the API server is configured to trust.
There are three ways for you to generate the required TLS certificate.

   - The easiest and default method is to install the [cert-manager](https://cert-manager.io/docs/installation/). With this, cert-manager will generate a self-signed certificate. 
   - Second way is to provide your own issuer by configuring the `mutatingWebhook.certManager.issuerRef` value. You need to spcify the kind (Issuer or ClusterIssuer) and the name. This method also requires cert-manager
   - Last way is to manually create the tls secret in the same namespace where lm-webhook will be deployed. In this case, you need to set `mutatingWebhook.certManager.enabled` to false.
     
     - Create the namespace for the lm-webhook if not exists 
       ```
       kubectl create namespace lm-webhook
       ```

     - Create the tls secret in the created namespace
       ```
        kubectl create secret tls lm-webhook-tls-cert \
          --cert=path/to/cert/file \
          --key=path/to/key/file \
          -n lm-webhook
       ```

       or you can also create tls secret by applying the following secret configuration
      
        ```
          kubectl apply -f - <<EOF
          apiVersion: v1
          kind: Secret
          metadata:
            name: lm-webhook-tls-cert
            namespace: lm-webhook
          type: kubernetes.io/tls
          data:
            tls.crt: |
              # your signed cert
            tls.key: |
              # your private key
          EOF
        ```
      - Set the base64 encoded value of CA trust chain to the `mutatingWebhook.caBundle`, which will be used by the api-server to validate the tls certificates. 

        **Note:** Default tls secret name used in lm-webhook is lm-webhook-tls-cert. If you are using different name, then you need to pass it by configuring the value of the `mutatingWebhook.tlsCertSecretName`
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

You can refer the following demo commands for deploying the lm-webhook with the helm-chart.
Run following command in the bash terminal from the helm-chart/lm-webhook directory.

- Using default tls certificate handling (using cert-manager)
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

- Using custom issuer other than self-signed issuer
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
    --set mutatingWebhook.certManager.issuerRef.name=private-ca-issuer \
    --set-file lmconfig=<path_to_external_config> \
    lm-webhook .
    ```

- Using your own tls certificates
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
    --set mutatingWebhook.certManager.enabled=false \
    --set mutatingWebhook.caBundle=$(base64 /tmp/cert/ca.pem) \
    --set-file lmconfig=<path_to_external_config> \
    lm-webhook .
    ```
  

#### Required Values:

- **cluster_name (default: ""):** Name of the k8s cluster in which lm-webhook will be deployed.
- **mutatingWebhook.caBundle (default: ""):** Base64 encoded value of CA trust chain. Required if `mutatingWebhook.certManager.enabled` is set to false.

#### Optional Values:

- **mutatingWebhook.objectSelector (default: ""):** specifies the label based selectors for the objects (pod) for which the requests are required to be intercepted.
- **mutatingWebhook.namespaceSelector (default: ""):** specifies the label based selectors for the namespaces.
- **lmconfig (default: ""):** specifies the external config file path.
- **mutatingWebhook.tlsCertSecretName (default: ""):** tls secret name.
- **mutatingWebhook.certManager.issuerRef (default: ""):** custom issuer other than self-signed issuer.
- **loglevel (default: "debug"):** sets log level. Possible values are debug, info, error

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
      value: lmotel-svc:4317
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
#### External config hot-reload
External config file content can be modified by updating the configmap, which causes lm-webhook to reload the external config inside the container without pod restart.
**Note:** lm-webhook does not support real-time config reload. As the official Kubernetes documentation says, the total delay from the moment when the ConfigMap is updated to the moment when new keys are projected to the Pod can be as long as the kubelet sync period + cache propagation delay, where the cache propagation delay depends on the chosen cache type (it equals to watch propagation delay, ttl of cache, or zero correspondingly). 
So, it can take few seconds to reflect the updated configuration in the pod.

-----
### License

 This Source Code is subject to the terms of the Mozilla Public
  License, v. 2.0. If a copy of the MPL was not distributed with this
  file, You can obtain one at https://mozilla.org/MPL/2.0/.
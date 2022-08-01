---
title: "Examples"
draft: false
---

You can refer the following example commands for different scenarios for deploying the lm-k8s-webhook with the helm-chart.

> Note: You should check the [troubleshooting guide](https://logicmonitor.github.io/lm-k8s-webhook/troubleshooting-guide) in case you face any issue in the deployment of the lm-k8s-webhook.

1. Using default tls certificate handling (using cert-manager)

    ```bash 
    $ helm install --debug --wait -n lm-k8s-webhook \
    --create-namespace \
    --set cluster_name="your-k8s-cluster-name" \
    lm-k8s-webhook .
    ```
---
2. Using custom issuer other than self-signed issuer

    ```bash
    $ helm install --debug --wait -n lm-k8s-webhook \
    --create-namespace \
    --set cluster_name="your-k8s-cluster-name" \
    --set mutatingWebhook.certManager.issuerRef.name=private-ca-issuer \
    --set mutatingWebhook.certManager.issuerRef.kind=Issuer \
    lm-k8s-webhook .
    ```
---
3. Using your own tls certificates

    ```bash
    $ helm install --debug --wait -n lm-k8s-webhook \
    --create-namespace \
    --set cluster_name="your-k8s-cluster-name" \
    --set mutatingWebhook.certManager.enabled=false \
    --set mutatingWebhook.caBundle=$(base64 /tmp/cert/ca.pem) \
    lm-k8s-webhook .
    ```
---
4. Using ObjectSelector and NamespaceSelector
    
    * ObjectSelector used here is:

    ```yaml
    objectSelector:
      matchLabels:
        tier: backend
      matchExpressions:
       - key: type
         operator: In
         values: ["application","service"]
    ```

    * NamespaceSelector used here is:

    ```yaml
    namespaceSelector:
      matchExpressions:
      - key: environment
        operator: In
        values: ["dev","staging"]
    ``` 

    * Corresponding helm command will look like:

    ```bash
    $ helm install --debug --wait -n lm-k8s-webhook \
    --create-namespace \
    --set cluster_name="your-k8s-cluster-name" \
    --set mutatingWebhook.objectSelector.matchLabels.tier="backend" \
    --set mutatingWebhook.objectSelector.matchExpressions[0].key="type" \
    --set mutatingWebhook.objectSelector.matchExpressions[0].operator="In" \
    --set mutatingWebhook.objectSelector.matchExpressions[0].values[0]=application \
    --set mutatingWebhook.objectSelector.matchExpressions[0].values[1]=service \
    --set mutatingWebhook.namespaceSelector.matchExpressions[0].key="environment" \
    --set mutatingWebhook.namespaceSelector.matchExpressions[0].operator="In" \
    --set mutatingWebhook.namespaceSelector.matchExpressions[0].values[0]="dev" \
    --set mutatingWebhook.namespaceSelector.matchExpressions[0].values[1]="staging" \
    lm-k8s-webhook .
    ```
---
5. Using external configuration

    ```bash
    $ helm install --debug --wait -n lm-k8s-webhook \
    --create-namespace \
    --set cluster_name="your-k8s-cluster-name" \
    --set-file lmK8sWebhook.config=<path_to_external_config> \
    lm-k8s-webhook .
    ```
---

6. Enabling lm-config-reloader by passing the lm-config-reloader config

    ```bash
    $ helm install --debug --wait -n lm-k8s-webhook \
    --create-namespace \
    --set cluster_name="your-k8s-cluster-name" \
    --set-file lmK8sWebhook.config=<path_to_external_config> \
    --set-file lmConfigReloader.config=<path_to_lm_reloader_config> \
    lm-k8s-webhook .
    ```
---
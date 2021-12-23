---
title: "Examples"
draft: false
---

You can refer the following example commands for different scenarios for deploying the lm-webhook with the helm-chart.

> Note: You should check the [troubleshooting guide](https://logicmonitor.github.io/lm-k8s-webhook/troubleshooting-guide) in case you face any issue in the deployment of the lm-webhook.

1. Using default tls certificate handling (using cert-manager)

    ```bash 
    $ helm install --debug --wait -n lm-webhook \
    --create-namespace \
    --set cluster_name="your-k8s-cluster-name" \
    lm-webhook .
    ```
---
2. Using custom issuer other than self-signed issuer

    ```bash
    $ helm install --debug --wait -n lm-webhook \
    --create-namespace \
    --set cluster_name="your-k8s-cluster-name" \
    --set mutatingWebhook.certManager.issuerRef.name=private-ca-issuer \
    --set mutatingWebhook.certManager.issuerRef.kind=Issuer \
    lm-webhook .
    ```
---
3. Using your own tls certificates

    ```bash
    $ helm install --debug --wait -n lm-webhook \
    --create-namespace \
    --set cluster_name="your-k8s-cluster-name" \
    --set mutatingWebhook.certManager.enabled=false \
    --set mutatingWebhook.caBundle=$(base64 /tmp/cert/ca.pem) \
    lm-webhook .
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
    $ helm install --debug --wait -n lm-webhook \
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
    lm-webhook .
    ```
---
5. Using external configuration

    ```bash
    $ helm install --debug --wait -n lm-webhook \
    --create-namespace \
    --set cluster_name="your-k8s-cluster-name" \
    --set-file lmconfig=<path_to_external_config> \
    lm-webhook .
    ```
---
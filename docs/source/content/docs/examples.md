---
title: "Examples"
draft: false
menu:
  main:
    parent: Docs
    identifier: "Examples"
    weight: 5
---

You can refer the following example commands for different scenarios for deploying the lm-webhook with the helm-chart.

- Using default tls certificate handling (using cert-manager)

    ```
    $ helm install --debug --wait -n lm-webhook \
    --create-namespace \
    --set cluster_name="your-k8s-cluster-name" \
    lm-webhook .
    ```

- Using custom issuer other than self-signed issuer

    ```
    $ helm install --debug --wait -n lm-webhook \
    --create-namespace \
    --set cluster_name="your-k8s-cluster-name" \
    --set mutatingWebhook.certManager.issuerRef.name=private-ca-issuer \
    --set mutatingWebhook.certManager.issuerRef.kind=Issuer \
    lm-webhook .
    ```

- Using your own tls certificates

    ```
    $ helm install --debug --wait -n lm-webhook \
    --create-namespace \
    --set cluster_name="your-k8s-cluster-name" \
    --set mutatingWebhook.certManager.enabled=false \
    --set mutatingWebhook.caBundle=$(base64 /tmp/cert/ca.pem) \
    lm-webhook .
    ```
- Using ObjectSelector and NamespaceSelector
    
    * ObjectSelector used here is:

    ```
    objectSelector:
      matchLabels:
        tier: backend
      matchExpressions:
       - key: type
         operator: In
         values: ["application","service"]
    ```

    * NamespaceSelector used here is:

    ```
    namespaceSelector:
      matchExpressions:
      - key: environment
        operator: In
        values: ["dev","staging"]
    ``` 

    * Corresponding helm command will look like:

    ```
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

- Using external configuration

    ```
    $ helm install --debug --wait -n lm-webhook \
    --create-namespace \
    --set cluster_name="your-k8s-cluster-name" \
    --set-file lmconfig=<path_to_external_config> \
    lm-webhook .
    ```
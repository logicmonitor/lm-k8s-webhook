---
title: "Helm chart configurations"
draft: false
menu:
  main:
    parent: Configurations
    identifier: "Helm chart configurations"
    weight: 3
---

## Required Values

- **cluster_name (default: ""):** Name of the k8s cluster in which lm-webhook will be deployed.
- **mutatingWebhook.caBundle (default: ""):** Base64 encoded value of CA trust chain. Required if `mutatingWebhook.certManager.enabled` is set to false.
---
## Optional Values

- **mutatingWebhook.objectSelector (default: ""):** specifies the label based selectors for the objects (pod) for which the requests are required to be intercepted.
- **mutatingWebhook.namespaceSelector (default: ""):** specifies the label based selectors for the namespaces.
- **mutatingWebhook.failurePolicy (default: "Ignore"):** Allowed values are Ignore or Fail. Ignore means that an error calling the webhook is ignored and the API request is allowed to continue. Fail means that an error calling the webhook causes the admission to fail and the API request to be rejected.
- **mutatingWebhook.timeoutSeconds (default: 30)** Timeout for webhook call in seconds.
> Note: Default timeout for a webhook call is 10 seconds for webhooks registered created using `admissionregistration.k8s.io/v1`, and 30 seconds for webhooks created using `admissionregistration.k8s.io/v1beta1`. Starting in kubernetes 1.14 you can set the timeout and it is encouraged to use a small timeout for webhooks.
- **mutatingWebhook.tlsCertSecretName (default: ""):** tls secret name.
- **mutatingWebhook.certManager.issuerRef (default: ""):** custom issuer other than self-signed issuer.
- **mutatingWebhook.certManager.enabled (default: true):** Allows cert-manager to manage the lm-webhook's tls certificates. Please make it false if you want to generate & manage tls certificates for the lm-webhook on your own.
- **lmwebhook.config (default: ""):** specifies the external config file path.
- **lmwebhook.loglevel (default: "debug"):** sets log level. Possible values are debug, info, error
- **lmwebhook.image.repository (default: "docker-registry.logicmonitor.net/development/lm-webhook")** The image respository of the lm-k8s-webhook container.
- **lmwebhook.image.tag:** The image tag of lm-k8s-webhook
- **lmwebhook.image.pullPolicy (default: "Always"):** The image pull policy of the lm-k8s-webhook container.
- **lmwebhook.imagePullSecrets:** The docker secret to pull the lm-k8s-webhook image
- **service.name (default: lm-webhook-svc):** Service name of the lm-k8s-webhook
- **service.port (default: 443):** Service Port of the lm-k8s-webhook
- **tolerations (default: []):** Tolerations are applied to pods, and allow the pods to schedule onto nodes with matching taints.
---
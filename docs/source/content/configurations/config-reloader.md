---
title: "LM Config Reloader (Optional)"
draft: false
menu:
  main:
    parent: Configurations
    identifier: "LM Config Reloader"
    weight: 3
---

## Overview
- LM-K8s-Webhook uses 2 important configurations, one is [Additional attribute config](https://logicmonitor.github.io/lm-k8s-webhook/configurations/additional-attributes-config/) and other one is `lm-k8s-webhook-mutating-webhook-configuration` which is of kind [MutatingWebhookConfiguration](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#mutatingwebhookconfiguration-namespaceselector-1-0) which basically contains the information about how the mutating webhook should be configured. 
- You can have these two configuration files in your Github repo and once you update these configurations on Github, `lm-config-reloader` will fetch the updated configurations and update the configurations in the container. 

---
## Configurations
By default, `lm-config-reloader` is disabled. You need to follow following steps to configure `lm-config-reloader`:
- As a part of MutatingWebhookConfiguration i.e. `lm-k8s-webhook-mutating-webhook-configuration`, as of now you can only update the __ObjectSelector__ and __NamespaceSelector__ using `LM-K8s-Reloader`. You can refer to the file content that is shown below and create your own MutatingWebhookConfiguration file by replacing the ObjectSelector and NamespaceSelector with your selectors and push it to the Github repo. 

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: lm-k8s-webhook-mutating-webhook-configuration
webhooks:
- name: lm-k8s-webhook-svc.lm-k8s-webhook.svc.cluster.local
  objectSelector:
  matchLabels:
    lm-k8s-webhook: enabled
  namespaceSelector:
    matchExpressions:
    - key: environment
      operator: In
      values:
      - dev
```

- For the `additional attribute config` file, you can refer to [Additional attribute config](https://logicmonitor.github.io/lm-k8s-webhook/configurations/additional-attributes-config/) section. Once config is ready, push this file to your Github repo. 
- Get the `lm-config-reloader` configuration file from [here](https://github.com/logicmonitor/lm-k8s-webhook/blob/main/lm-config-reloader/reloader-config.yaml) and update this config file to include the details of your Github repo where configurations are placed. 
- You need to pass this config file while deploying the LM-K8s-Webhook. You can refer to the example 6 from [example](https://logicmonitor.github.io/lm-k8s-webhook/examples/) section.

---
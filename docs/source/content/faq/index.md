---
title: "FAQ's"
draft: false
---

**1. Does `LM-K8s-Webhook` support hot-reloading of the external configuration passed to it ?**
* Yes, external config file content can be modified by updating the configmap, which causes lm-k8s-webhook to reload the external config inside the container without pod restart.
> **Note:** lm-k8s-webhook does not support real-time config reload. As the official Kubernetes documentation says, the total delay from the moment when the ConfigMap is updated to the moment when new keys are projected to the Pod can be as long as the kubelet sync period + cache propagation delay, where the cache propagation delay depends on the chosen cache type (it equals to watch propagation delay, ttl of cache, or zero correspondingly). 
So, it can take few seconds to reflect the updated configuration in the pod.

---
**2. Do I need to make any changes in application pods to make use of the `LM-K8s-Webhook` ?**
* If you have configured selectors i.e. `Object selector`, `Namespace selector` while deploying the LM-K8s-Webhook, then you need to make sure that your pods and namespace satisfy corresponding selectors.
---
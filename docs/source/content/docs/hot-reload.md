---
title: "Hot-reload"
draft: false
menu:
  main:
    parent: Docs
    identifier: "Hot-reload"
    weight: 3
---

- External config file content can be modified by updating the configmap, which causes lm-webhook to reload the external config inside the container without pod restart.
> **Note:** lm-webhook does not support real-time config reload. As the official Kubernetes documentation says, the total delay from the moment when the ConfigMap is updated to the moment when new keys are projected to the Pod can be as long as the kubelet sync period + cache propagation delay, where the cache propagation delay depends on the chosen cache type (it equals to watch propagation delay, ttl of cache, or zero correspondingly). 
So, it can take few seconds to reflect the updated configuration in the pod.
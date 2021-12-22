---
title: "LM-K8s-Webhook"
draft: false
type: index
---


**LM-Webhook** is the implementation of the `Kubernetes Mutating Admission webhook`. Some of the key features of the LM-Webhook are:

- LM-Webhook can be used to inject the kubernetes specific resource attributes like pod name, ip, pod namespace, service namespace, pod UUID in the pod as an environment variable, which avoids the need of manually updating the deployment manifests to include these resource attributes. 
- Custom environment variables can also be injected by passing the external configuration.
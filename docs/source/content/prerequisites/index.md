---
title: "Prerequisites"
draft: false
---

1. Ensure that Kubernetes cluster is at least as new as v1.16 ( to use `admissionregistration.k8s.io/v1` ) or v1.9 ( to use `admissionregistration.k8s.io/v1beta1` )
---
2. Ensure that `MutatingAdmissionWebhook admission controller` is enabled.
You can check if it is enabled by looking at the admission plugins that are enabled by running the following command in `kube-apiserver`:
   ```bash
     $ kube-apiserver -h | grep enable-admission-plugins
   ```
   If not enabled, then you can enable this plugin by running the following command in `kube-apiserver`:
   ```bash
     $ kube-apiserver --enable-admission-plugins=MutatingAdmissionWebhook  
   ```
---
3. Ensure that the API (`admissionregistration.k8s.io/v1` or `admissionregistration.k8s.io/v1beta1` depending upon the k8s version being used) is enabled by using the following commands: 

   * For `admissionregistration.k8s.io/v1beta1` API:
   ```bash
   $ kubectl api-versions | grep admissionregistration.k8s.io/v1beta1
   ```

   Output should be:

   ```
     admissionregistration.k8s.io/v1beta1
   ```
   
   * For `admissionregistration.k8s.io/v1` API: 
   ```bash
     $ kubectl api-versions | grep admissionregistration.k8s.io/v1
   ```
   Output should be:

   ```
     admissionregistration.k8s.io/v1
   ```
---
4. TLS Certificate Requirement:

In Kubernetes, in order for the API server to communicate with the webhook component, the webhook requires a TLS certificate that the API server is configured to trust. You can refer to the [deployment section](https://logicmonitor.github.io/lm-k8s-webhook/deployment) to understand more about it.

---
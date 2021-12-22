---
title: "Getting Started"
draft: false
---

## Prerequisites:
* Ensure that Kubernetes cluster is at least as new as v1.16 ( to use `admissionregistration.k8s.io/v1` ) or v1.9 ( to use `admissionregistration.k8s.io/v1beta1` )
* Ensure that `MutatingAdmissionWebhook admission controller` is enabled.
You can check if it is enabled by looking at the admission plugins that are enabled by running the following command in `kube-apiserver`:
```
   $ kube-apiserver -h | grep enable-admission-plugins
```
   If not enabled, then you can enable this plugin by running the following command in `kube-apiserver`:
```
   $ kube-apiserver --enable-admission-plugins=MutatingAdmissionWebhook  
```
* Ensure that the API (`admissionregistration.k8s.io/v1` or `admissionregistration.k8s.io/v1beta1` depending upon the k8s version being used) is enabled by using the following commands: 

   * For `admissionregistration.k8s.io/v1beta1` API:
```
  $ kubectl api-versions | grep admissionregistration.k8s.io/v1beta1
```
  Output should be:
```
  admissionregistration.k8s.io/v1beta1
```
 
   * For `admissionregistration.k8s.io/v1` API: 
```
   $ kubectl api-versions | grep admissionregistration.k8s.io/v1
```
   Output should be:
```
   admissionregistration.k8s.io/v1
```
* **TLS Certificate Requirement:**

In Kubernetes, in order for the API server to communicate with the webhook component, the webhook requires a TLS certificate that the API server is configured to trust. Following are some of the ways in which you can configure the required TLS certificate.

**Option 1**:
The easiest and default method is to install the [cert-manager](https://cert-manager.io/docs/installation/). With this option, cert-manager will generate a self-signed certificate. 
> Note: Please make sure that the components (pods) deployed by the cert-manager are up and running. By default cert-manager deploys its pods in the `cert-manager` namespace.

**Option 2**:
If you want to use the cert-manager but want to use your own issuer to generate the certificates, then you can go with this option.  

For that you need to set the `mutatingWebhook.certManager.issuerRef.name` property of the lm-webhook helm chart to the name of the issuer which is deployed in your Kubernetes cluster and you also need to set the `mutatingWebhook.certManager.issuerRef.kind` property to the kind of the issuer (Issuer or ClusterIssuer).

Both option 1 and option 2 need a cert-manager installed in your k8s cluster.
> Note: Please make sure that the components (pods) deployed by the cert-manager are up and running. By default cert-manager deploys its pods in the `cert-manager` namespace.

**Option 3**:
If you want to generate & manage tls certificates for the lm-webhook on your own, you can create the required certificate and key for the lm-webhook and manually create the tls secret in the same namespace where lm-webhook will be deployed. 

In this case, you need to set `mutatingWebhook.certManager.enabled` to false, so that you don't need to set up cert-manager.

> Note: By default the service name of the lm-webhook is `lm-webhook-svc`. So, the server cert must be valid for `<svc_name>.<svc_namespace>.svc`

If you are following `option 3`, then once you have the required certificate and the key files ready for lm-webhook you can follow below steps:

1. Create the namespace for the lm-webhook if not exists

```
$ kubectl create namespace lm-webhook
```
2. Create the tls secret in the same namespace

Default tls secret name consumed in the lm-webhook is `lm-webhook-tls-cert`. If you are using different name, then you need to pass it by configuring the value of the `mutatingWebhook.tlsCertSecretName`

```
 $ kubectl create secret tls lm-webhook-tls-cert \
   --cert=path/to/cert/file \
   --key=path/to/key/file \
   -n lm-webhook
 ```

3. Set the base64 encoded value of the CA trust chain to the `mutatingWebhook.caBundle` which will be used by the api-server to validate the tls certificates.
 

## Deploying the lm-webhook helm-chart
* Depending on the certificate management you are using and the lm-webhook components like [selectors](https://logicmonitor.github.io/lm-k8s-webhook/docs/selectors/) and [external-config](https://logicmonitor.github.io/lm-k8s-webhook/docs/external-config/), you need to modify the helm command for the lm-webhook deployment. You can refer the [examples page](https://logicmonitor.github.io/lm-k8s-webhook/docs/examples/).

* For all the possible values that can be configured with lm-webhook helm chart refer to [configuration page](https://logicmonitor.github.io/lm-k8s-webhook/docs/configuration/)
* Helm chart for the lm-webhook is available at https://github.com/logicmonitor/lm-k8s-webhook/tree/main/helm-chart/lm-webhook path.
* The simplest lm-webhook deployment without passing any selectors and external configuration can be done by running the following command in bash terminal from the `helm-chart/lm-webhook` directory.

```
$ helm install --debug --wait -n lm-webhook \
--create-namespace \
--set cluster_name="<cluster_name>" \
lm-webhook .
```
 
## Deploying the application pods
* Once the lm-webhook is up and running, you can deploy the application pods that you wanted to get mutated. 
* If you have configured selectors i.e. Object selector, Namespace selector then you need to make sure that your pods and namespace should satisfy corresponding selectors. 
* If everything goes well, then after the pod gets deployed, you can see that pod has the Kubernetes resource attributes as an environment variables injected into it. 

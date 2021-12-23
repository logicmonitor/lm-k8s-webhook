---
title: "Selectors"
draft: false
menu:
  main:
    parent: Configurations
    identifier: "Selectors"
    weight: 1
---

Selectors can be used to limit which requests can be intercepted by the webhook based on the labels.
Two types of selectors can be specified in _MutatingWebhookConfiguration_  i.e. _ObjectSelector_ and _NamespaceSelector_.

Both ObjectSelector and NamespaceSelector can use matchLabels and matchExpressions to specify the selectors.
You can check [working with kubernetes objects and labels](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) for more details.

---

## ObjectSelector

`ObjectSelector` is used to specify the label based selectors for the objects (pod) for which the requests are required to be intercepted.

**Example:** 
Using matchLabels, objectSelector can be specified as follows:

```yaml
 objectSelector:
    matchLabels:
      tier: backend
```

With this selectors the requests for objects (pod) with label tier = backend will be intercepted.      
Using matchExpressions, objectSelector can be specified as follows:

```yaml
objectSelector:
  matchExpressions:
    - key: tier
      operator: In
      values: ["frontend","backend"]
```

With this selectors the requests for objects (pod) with label tier = backend or tier = frontend will be intercepted.

---

## NamespaceSelector

`NamespaceSelector` is used to specify the label based selectors for the namespaces.

**Example:** 
Using matchLabels, namespaceSelector can be specified as follows:

```yaml
 namespaceSelector:
    matchLabels:
      environment: development
```
With this selectors the requests for objects (pods) with label tier = backend will be intercepted.      
Using matchExpressions, namespaceSelector can be specified as follows:

```yaml
namespaceSelector:
  matchExpressions:
    - key: tier
      operator: In
      values: ["development","staging"]
```

> Note: If you have configured selectors i.e. `Object selector` & `Namespace selector` while deploying the `LM-K8s-Webhook`, then you need to make sure that your pods and corresponding namespace satisfy the configured selectors.

* You can check the [examples page](https://logicmonitor.github.io/lm-k8s-webhook/docs/examples/) to get an idea of using these selectors.
---
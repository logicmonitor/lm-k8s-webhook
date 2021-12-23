---
title: "Additional attributes config"
draft: false
menu:
  main:
    parent: Configurations
    identifier: "Additional attributes config"
    weight: 2
---

Currently as a part of the external config, user can define the custom environment variables that are to be injected into the application pods.

You can download the sample external config file from here: https://github.com/logicmonitor/lm-k8s-webhook/blob/main/sampleconfig.yaml

**Example:**
```yaml
  lmEnvVars:
    resource:
      - env: 
          name: SERVICE_ACCOUNT_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.serviceAccountName
        resAttrName: serviceaccount.name
        overrideDisabled: true
      - env:
          name: SERVICE_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.labels['app-namespace']
      - env:
          name: SERVICE_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.labels['app-name']
    operation:
      - env:
          name: COMPANY_NAME
          value: ABC Corporation
        overrideDisabled: true 
      - env:
          name: OTLP_ENDPOINT
          value: lmotel-svc:4317
        overrideDisabled: true 
      - env:
          name: OTEL_JAVAAGENT_ENABLED
          value: true
        overrideDisabled: true
      - env:
          name: DEPLOYMENT_ENV
          value: production
```

environment variables can be of two types, i.e. `resource` and `operation`
- `Resource` holds the resource environment variables, which will be the part of _OTEL_RESOURCE_ATTRIBUTES_.
- `Operation` holds the operation environment variables, which will not be the part of _OTEL_RESOURCE_ATTRIBUTES_ but can be used in the application for custom use cases.
- `resAttrName` field can be used only in resource section. The value assigned to resAttrName will be used as a name of the resource attribute instead of the actual env variable name while passing it through the _OTEL_RESOURCE_ATTRIBUTES_.
- `overrideDisabled` field can be used in both, resource as well as operation sections. It decides if the value of the env variable which is defined in external config is allowed to be overriden by the same name env variable from the container definition. Default value of this field is false, which means that overriding of the value of the env variable defined in external config is allowed.

* LM-Webhook injects following environment variables in the application pods. 

| SR. No. | Environment Variable Name | 
| ---: | :--- | 
| 1 | LM_APM_CLUSTER_NAME | 
| 2 | LM_APM_NODE_NAME |
| 3 | LM_APM_POD_NAME |
| 4 | LM_APM_POD_IP | 
| 5 | LM_APM_POD_NAMESPACE | 
| 6 | LM_APM_POD_UID | 
| 7 | SERVICE_NAMESPACE | 
| 8 | SERVICE_NAME |
| 9 | OTEL_RESOURCE_ATTRIBUTES | 

* It is not recommanded to explicitely specify these environment variables except `SERVICE_NAME`, `SERVICE_NAMESPACE` & `OTEL_RESOURCE_ATTRIBUTES` as a part of pod definition. 
Default value of `SERVICE_NAMESPACE` is the value of the pod namespace, which can be overriden, either by specifying it as a part of pod definition (if overriding is allowed) or in the external configuration. 

* You can pass the resource attributes which are not getting set by the lm-webhook by defining the `OTEL_RESOURCE_ATTRIBUTES` env variable in the pod definition, which will get merged with the ones which are defined by lm-webhook.

* Values for `SERVICE_NAME` and `SERVICE_NAMESPACE` can also be specified in terms of pod label as shown in above example config. So that value of the specified pod label can be used as a `SERVICE_NAME` or `SERVICE_NAMESPACE`.
---
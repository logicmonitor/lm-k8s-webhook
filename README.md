
# LM-K8s-Webhook

[![codecov](https://codecov.io/gh/logicmonitor/lm-k8s-webhook/branch/main/graph/badge.svg?token=DTWHXaXZzl)](https://codecov.io/gh/logicmonitor/lm-k8s-webhook)
[![build_and_test](https://github.com/logicmonitor/lm-k8s-webhook/actions/workflows/continuous-integration.yml/badge.svg)](https://github.com/logicmonitor/lm-k8s-webhook/actions/workflows/continuous-integration.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/logicmonitor/lm-k8s-webhook.svg)](https://pkg.go.dev/github.com/logicmonitor/lm-k8s-webhook)
[![Go Report Card](https://goreportcard.com/badge/github.com/logicmonitor/lm-k8s-webhook)](https://goreportcard.com/report/github.com/logicmonitor/lm-k8s-webhook)

## Overview

**LM-K8s-Webhook** is the implementation of the [Kubernetes Mutating Admission webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/). Some of the key features of the `LM-K8s-Webhook` are:

- LM-K8s-Webhook can be used to inject the kubernetes specific resource attributes like pod name, ip, pod namespace, service namespace, pod UUID in the pod as an environment variable, which avoids the need of manually updating the deployment manifests to include these resource attributes. 
- Custom environment variables can also be injected by passing the external configuration.  

## Getting started

See the [Getting Started](https://logicmonitor.github.io/lm-k8s-webhook/) document.

## Troubleshooting

If you encounter issues, review the [troubleshooting docs](https://logicmonitor.github.io/lm-k8s-webhook/troubleshooting-guide/)

## License

[![license](https://img.shields.io/github/license/logicmonitor/lm-k8s-webhook.svg)](https://github.com/logicmonitor/lm-k8s-webhook/blob/main/LICENSE)
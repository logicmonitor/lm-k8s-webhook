---
title: "Troubleshooting Guide"
draft: false
---

1. If you are using `zsh terminal` and you are using [] notation in the helm chart deployment command, then you might encounter an error saying `zsh: no matches found:`. [] syntax has its meaning in zsh. 
So there are two simple ways to step aside.
    * **Change to bash:** switch to bash by just entering `bash`. And then run you helm install again.
    * **noglob:** you can use noglob. 
        ```bash
            $ noglob helm install --debug --wait -n lm-webhook \
            --create-namespace \
            --set cluster_name="your-k8s-cluster-name" \
            --set mutatingWebhook.objectSelector.matchLabels.tier="backend" \
            --set mutatingWebhook.objectSelector.matchExpressions[0].key="type" \
            --set mutatingWebhook.objectSelector.matchExpressions[0].operator="In" \
            --set mutatingWebhook.objectSelector.matchExpressions[0].values[0]=application \
            --set mutatingWebhook.objectSelector.matchExpressions[0].values[1]=service \
            --set mutatingWebhook.namespaceSelector.matchExpressions[0].key="environment" \
            --set mutatingWebhook.namespaceSelector.matchExpressions[0].operator="In" \
            --set mutatingWebhook.namespaceSelector.matchExpressions[0].values[0]="dev" \
            --set mutatingWebhook.namespaceSelector.matchExpressions[0].values[1]="staging" \
            lm-webhook .
        ```
---
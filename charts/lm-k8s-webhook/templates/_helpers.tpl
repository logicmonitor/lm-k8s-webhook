{{/*
Expand the name of the chart.
*/}}
{{- define "lm-k8s-webhook.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "lm-k8s-webhook.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "lm-k8s-webhook.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "lm-k8s-webhook.labels" -}}
helm.sh/chart: {{ include "lm-k8s-webhook.chart" . }}
{{ include "lm-k8s-webhook.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "lm-k8s-webhook.selectorLabels" -}}
app.kubernetes.io/name: {{ include "lm-k8s-webhook.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "lm-k8s-webhook.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "lm-k8s-webhook.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Return the appropriate apiVersion for rbac.
*/}}
{{- define "rbac.apiVersion" -}}
{{- if .Capabilities.APIVersions.Has "rbac.authorization.k8s.io/v1" }}
{{- print "rbac.authorization.k8s.io/v1" -}}
{{- else -}}
{{- print "rbac.authorization.k8s.io/v1beta1" -}}
{{- end -}}
{{- end -}}

{{/*
Return the appropriate apiVersion for admissionregistration.
*/}}
{{- define "admissionregistration.apiVersion" -}}
{{- if .Capabilities.APIVersions.Has "admissionregistration.k8s.io/v1" }}
{{- print "admissionregistration.k8s.io/v1" -}}
{{- else -}}
{{- print "admissionregistration.k8s.io/v1beta1" -}}
{{- end -}}
{{- end -}}

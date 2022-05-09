{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "bcs-argocd-server.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "bcs-argocd-server.fullname" -}}
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
{{- define "bcs-argocd-server.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "bcs-argocd-server.labels" -}}
helm.sh/chart: {{ include "bcs-argocd-server.chart" . }}
{{ include "bcs-argocd-server.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "bcs-argocd-server.selectorLabels" -}}
app.kubernetes.io/platform: bk-bcs
app.kubernetes.io/name: {{ include "bcs-argocd-server.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "bcs-argocd-server.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "bcs-argocd-server.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create a default fully qualified app name for etcd subchart
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "bcs-argocd-server.etcd.fullname" -}}
{{- if .Values.global.etcd.fullnameOverride -}}
{{- .Values.global.etcd.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default "etcd" .Values.global.etcd.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Return the Etcd Address
*/}}
{{- define "bcs-argocd-server.etcd.address" -}}
{{- if .Values.global.etcd.enabled }}
    {{- printf "http://%s:2379" (include "bcs-argocd-server.etcd.fullname" .) -}}
{{- else -}}
    {{- printf "%s" .Values.externalEtcd.address -}}
{{- end -}}
{{- end -}}

{{/*
Return the Etcd CA
*/}}
{{- define "bcs-argocd-server.etcd.ca" -}}
{{- if .Values.global.etcd.enabled }}
    {{- printf "%s" .Values.global.etcd.auth.client.caFilename -}}
{{- else -}}
    {{- printf "%s" .Values.externalEtcd.ca -}}
{{- end -}}
{{- end -}}

{{/*
Return the Etcd Cert
*/}}
{{- define "bcs-argocd-server.etcd.cert" -}}
{{- if .Values.global.etcd.enabled }}
    {{- printf "%s" .Values.global.etcd.auth.client.certFilename -}}
{{- else -}}
    {{- printf "%s" .Values.externalEtcd.cert -}}
{{- end -}}
{{- end -}}

{{/*
Return the Etcd Key
*/}}
{{- define "bcs-argocd-server.etcd.key" -}}
{{- if .Values.global.etcd.enabled }}
    {{- printf "%s" .Values.global.etcd.auth.client.certKeyFilename -}}
{{- else -}}
    {{- printf "%s" .Values.externalEtcd.key -}}
{{- end -}}
{{- end -}}
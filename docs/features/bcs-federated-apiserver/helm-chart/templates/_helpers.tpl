{{/*
Expand the name of the chart.
*/}}
{{- define "bcs-federated-apiserver.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "bcs-federated-apiserver.fullname" -}}
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
{{- define "bcs-federated-apiserver.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "bcs-federated-apiserver.labels" -}}
helm.sh/chart: {{ include "bcs-federated-apiserver.chart" . }}
{{ include "bcs-federated-apiserver.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "bcs-federated-apiserver.selectorLabels" -}}
app.kubernetes.io/platform: bk-bcs
app.kubernetes.io/name: {{ include "bcs-federated-apiserver.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
etcd common labels
*/}}
{{- define "bcs-federated-apiserver-etcd.labels" -}}
helm.sh/chart: {{ include "bcs-federated-apiserver.chart" . }}
{{ include "bcs-federated-apiserver-etcd.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Etcd selector labels
*/}}
{{- define "bcs-federated-apiserver-etcd.selectorLabels" -}}
app.kubernetes.io/platform: bk-bcs
app.kubernetes.io/name: {{ .Values.etcd.name }}
app.kubernetes.io/instance: {{ .Values.etcd.name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "bcs-federated-apiserver.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "bcs-federated-apiserver.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

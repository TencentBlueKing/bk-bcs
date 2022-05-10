{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "bcs-data-manager.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "bcs-data-manager.fullname" -}}
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
{{- define "bcs-data-manager.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "bcs-data-manager.labels" -}}
helm.sh/chart: {{ include "bcs-data-manager.chart" . }}
{{ include "bcs-data-manager.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "bcs-data-manager.selectorLabels" -}}
app.kubernetes.io/platform: bk-bcs
app.kubernetes.io/name: {{ include "bcs-data-manager.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "bcs-data-manager.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "bcs-data-manager.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Return the peer token(uuid format) environment pair, format:
- name: somekey
  value: 
or
- name: somekey
  valueFrom:
    secretKeyRef:
      
Usage: {{ include "bcs-data-manager.peerToken" ( dict "root" . "envName" "somekey" ) }}
*/}}
{{- define "bcs-data-manager.peerToken" -}}
{{- if .root.Values.env.BK_BCS_bcsDataManagerPeerToken }}
- name: {{ .envName }}
  value: {{ .root.Values.env.BK_BCS_bcsDataManagerPeerToken | quote }}
{{- else }}
- name: {{ .envName }}
  valueFrom:
    secretKeyRef:
      name: {{ include "bcs-data-manager.fullname" .root }}
      key: data-manager-peer-token
{{- end }}
{{- end }}
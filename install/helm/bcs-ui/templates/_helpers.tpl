{{/*
Expand the name of the chart.
*/}}
{{- define "bcs-ui.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "bcs-ui.fullname" -}}
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
{{- define "bcs-ui.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "bcs-ui.labels" -}}
helm.sh/chart: {{ include "bcs-ui.chart" . }}
{{ include "bcs-ui.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "bcs-ui.selectorLabels" -}}
app.kubernetes.io/name: {{ include "bcs-ui.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "bcs-ui.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "bcs-ui.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{- define "bcs-ui.envs" -}}
- name: POD_IP
    valueFrom:
    fieldRef:
        fieldPath: status.podIP
- name: POD_IPs # ipv6双栈
    valueFrom:
    fieldRef:
        fieldPath: status.podIPs
{{- end }}

{{- define "bcs-ui.volumes" -}}
- name: POD_IP
    valueFrom:
    fieldRef:
        fieldPath: status.podIP
- name: POD_IPs # ipv6双栈
    valueFrom:
    fieldRef:
        fieldPath: status.podIPs
{{- end }}


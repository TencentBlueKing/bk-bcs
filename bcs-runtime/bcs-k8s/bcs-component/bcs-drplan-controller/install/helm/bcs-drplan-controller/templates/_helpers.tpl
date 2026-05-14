{{/*
Expand the name of the chart.
*/}}
{{- define "bcs-drplan-controller.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "bcs-drplan-controller.fullname" -}}
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
{{- define "bcs-drplan-controller.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "bcs-drplan-controller.labels" -}}
helm.sh/chart: {{ include "bcs-drplan-controller.chart" . }}
{{ include "bcs-drplan-controller.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "bcs-drplan-controller.selectorLabels" -}}
app.kubernetes.io/name: {{ include "bcs-drplan-controller.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "bcs-drplan-controller.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "bcs-drplan-controller.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the webhook service
*/}}
{{- define "bcs-drplan-controller.webhookServiceName" -}}
{{- printf "%s-webhook-service" (include "bcs-drplan-controller.fullname" .) }}
{{- end }}

{{/*
Create the webhook certificate secret name
*/}}
{{- define "bcs-drplan-controller.webhookCertSecretName" -}}
{{- printf "%s-webhook-cert" (include "bcs-drplan-controller.fullname" .) }}
{{- end }}

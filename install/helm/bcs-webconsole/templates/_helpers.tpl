{{/*
Expand the name of the chart.
*/}}
{{- define "bcs-webconsole.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "bcs-webconsole.fullname" -}}
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
{{- define "bcs-webconsole.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "bcs-webconsole.labels" -}}
helm.sh/chart: {{ include "bcs-webconsole.chart" . }}
{{ include "bcs-webconsole.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "bcs-webconsole.selectorLabels" -}}
app.kubernetes.io/name: {{ include "bcs-webconsole.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "bcs-webconsole.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "bcs-webconsole.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{- define "bcs-webconsole.envs" -}}
- name: POD_IPs # ipv6双栈
  valueFrom:
    fieldRef:
      fieldPath: status.podIPs
- name: POD_IP
  valueFrom:
    fieldRef:
      fieldPath: status.podIP
- name: BK_APP_CODE
  value: {{ .Values.svcConf.base_conf.app_code }}
- name: BK_APP_SECRET
  value: {{ .Values.svcConf.base_conf.app_secret }}
- name: BK_PAAS_HOST
  value: {{ .Values.svcConf.base_conf.bk_paas_host }}
- name: BK_IAM_HOST
  value: {{ .Values.svcConf.auth_conf.host }}
- name: REDIS_PASSWORD
  value: {{ .Values.svcConf.redis.password }}
- name: BCS_APIGW_TOKEN
  value: {{ .Values.svcConf.bcs_conf.token }}
- name: BCS_APIGW_PUBLIC_KEY
  value: {{ .Values.svcConf.bcs_conf.jwt_public_key }}
{{- range $key, $val := .Values.envs }}
- name: {{ $key }}
  value: {{ $val | quote }}
{{- end }}
{{- end }}

{{- define "bcs-webconsole.confName" -}}
{{- printf "%s-%s"  (include "bcs-webconsole.fullname" .) "conf" }}
{{- end }}

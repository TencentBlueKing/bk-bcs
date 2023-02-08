{{/*
Expand the name of the chart.
*/}}
{{- define "bcs-monitor.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "bcs-monitor.fullname" -}}
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
{{- define "bcs-monitor.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "bcs-monitor.labels" -}}
helm.sh/chart: {{ include "bcs-monitor.chart" . }}
{{ include "bcs-monitor.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "bcs-monitor.selectorLabels" -}}
app.kubernetes.io/name: {{ include "bcs-monitor.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "bcs-monitor.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "bcs-monitor.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}


{{- define "bcs-monitor.envs" -}}
- name: POD_IP
  valueFrom:
    fieldRef:
      fieldPath: status.podIP
- name: POD_IPs # ipv6双栈
  valueFrom:
    fieldRef:
      fieldPath: status.podIPs
- name: BK_SYSTEM_ID
  value: {{ .Values.global.bkAPP.systemID }}
- name: BK_APP_CODE
  value: {{ .Values.global.bkAPP.appCode }}
- name: BK_APP_SECRET
  value: {{ .Values.global.bkAPP.appSecret }}
- name: BK_PAAS_HOST
  value: {{ .Values.global.bkAPP.bkiamHost }}
- name: REDIS_PASSWORD
  value: {{ .Values.global.storage.redis.password }}
- name: BCS_APIGW_TOKEN
  value: {{ .Values.global.env.BK_BCS_gatewayToken}}
- name: BCS_APIGW_PUBLIC_KEY
  valueFrom:
    secretKeyRef:
      name: bcs-jwt
      key: public.key
- name: bcsEtcdHost
  value: "{{ include "bcs-common.etcd.host" ( dict "localStorage" .Values.storage "globalStorage" .Values.global.storage "namespace" .Release.Namespace ) }}"
- name: BKIAM_GATEWAY_SERVER
  value: {{ .Values.global.bkIAM.gateWayHost}}
{{- end }}

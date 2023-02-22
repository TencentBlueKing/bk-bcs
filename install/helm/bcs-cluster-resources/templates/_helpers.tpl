{{- define "bcs-cluster-resources.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "bcs-cluster-resources.fullname" -}}
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

{{- define "bcs-cluster-resources.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "bcs-cluster-resources.labels" -}}
helm.sh/chart: {{ include "bcs-cluster-resources.chart" . }}
{{ include "bcs-cluster-resources.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "bcs-cluster-resources.selectorLabels" -}}
app.kubernetes.io/name: {{ include "bcs-cluster-resources.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}


{{- define "bcs-cluster-resources.envs" -}}
- name: LOCAL_IP
  valueFrom:
    fieldRef:
      fieldPath: status.podIP
- name: POD_IPs # ipv6双栈
  valueFrom:
    fieldRef:
      fieldPath: status.podIPs
- name: BK_APP_CODE
  value: {{ .Values.global.bkAPP.appCode }}
- name: BK_APP_SECRET
  value: {{ .Values.global.bkAPP.appSecret }}
- name: BK_PAAS_HOST
  value: {{ .Values.global.bkIAM.bkiamHost }}
- name: BK_IAM_HOST
  value: {{ .Values.global.bkIAM.iamHost }}
- name: BK_IAM_GATEWAY_HOST
  value: {{ .Values.global.bkIAM.gateWayHost }}
- name: BK_IAM_SYSTEM_ID
  value: {{ .Values.global.bkAPP.systemID }}
- name: REDIS_PASSWORD
  value: {{ .Values.global.storage.redis.password }}
{{- range $key, $val := .Values.envs }}
- name: {{ $key }}
  value: {{ $val | quote }}
{{- end }}
{{- if .Values.svcConf.crGlobal.bcsApiGW.readAuthTokenFromEnv }}
- name: BCS_API_GW_AUTH_TOKEN
  valueFrom:
    secretKeyRef:
      name: bcs-password
      key: gateway_token
{{- end }}
{{- end }}

{{- define "bcs-cluster-resources.confCMName" -}}
{{- printf "%s-%s"  (include "bcs-cluster-resources.fullname" .) "conf" }}
{{- end }}

{{- define "bcs-cluster-resources.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "bcs-cluster-resources.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}


{{- define "envs" -}}
{{- range $key, $val := .Values.envs }}
- name: {{ $key }}
  value: {{ $val | quote }}
{{- end }}
- name: BK_IAM_RESOURCE_API_HOST
  value: {{ .Values.iam.BK_IAM_RESOURCE_API_HOST | default (printf "https://bcs-api-gateway.%s.svc.cluster.local" .Release.Namespace) }}
- name: BK_IAM_PROVIDER_PATH_PREFIX
  value: {{ .Values.iam.BK_IAM_PROVIDER_PATH_PREFIX | default "/bcsapi/v4/iam-provider" }}
- name: BCS_APP_APIGW_PUBLIC_KEY
  value: {{ index .Values.global.bkAPIGW "bcs-app" "bkApigatewayPublicKey" }}
- name: BK_REPO_TOKEN
  value: {{ include "app.bkRepoToken" . | quote }}
- name: BCS_APIGW_TOKEN
  valueFrom:
    secretKeyRef:
      name: bcs-password
      key: gateway_token
{{- end }}
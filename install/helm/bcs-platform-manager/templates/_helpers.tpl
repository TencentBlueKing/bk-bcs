{{- define "bcs-platform-manager.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "bcs-platform-manager.fullname" -}}
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

{{- define "bcs-platform-manager.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "bcs-platform-manager.labels" -}}
helm.sh/chart: {{ include "bcs-platform-manager.chart" . }}
{{ include "bcs-platform-manager.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "bcs-platform-manager.selectorLabels" -}}
app.kubernetes.io/name: {{ include "bcs-platform-manager.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "bcs-platform-manager.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.image "global" .Values.global) }}
{{- end -}}

{{- define "bcs-platform-manager.envs" -}}
- name: POD_IP
  valueFrom:
    fieldRef:
      fieldPath: status.podIP
- name: POD_IPs # ipv6双栈
  valueFrom:
    fieldRef:
      fieldPath: status.podIPs
- name: BCS_APIGW_PUBLIC_KEY
  valueFrom:
    secretKeyRef:
      key: public.key
      name: bcs-jwt
- name: BK_APP_CODE
  value: {{ .Values.global.bkAPP.appCode }}
- name: BK_APP_SECRET
  value: {{ .Values.global.bkAPP.appSecret }}
- name: BK_PAAS_HOST
  value: {{ .Values.global.bkIAM.bkiamHost }}
- name: BK_IAM_HOST
  value: {{ .Values.global.bkIAM.iamHost }}
- name: BKIAM_GATEWAY_SERVER
  value: {{ .Values.global.bkIAM.gateWayHost }}
- name: BK_SYSTEM_ID
  value: {{ .Values.global.bkAPP.systemID }}
{{- include "bcs-common.bcspwd.gatewayToken" ( dict "root" . "externalToken" .Values.global.env.BK_BCS_gatewayToken "envName" "BCS_APIGW_TOKEN" ) }}
{{- include "bcs-common.bcspwd.redis" ( dict "root" . "envName" "REDIS_PASSWORD" ) }}
{{- range $key, $val := .Values.envs }}
- name: {{ $key }}
  value: {{ $val | quote }}
{{- end }}
{{- if .Values.svcConf.bcs_conf.readAuthTokenFromEnv }}
- name: BCS_API_GW_AUTH_TOKEN
  valueFrom:
    secretKeyRef:
      name: bcs-password
      key: gateway_token
{{- end }}
- name: MONGO_ADDRESS
  value: "{{ include "bcs-common.mongodb.host" ( dict "localStorage" .Values.storage "externalMongo" .Values.svcConf.mongo.address "globalStorage" .Values.global.storage "namespace" .Release.Namespace ) }}"
- name: MONGO_USERNAME
  value: "{{ .Values.global.storage.mongodb.username | default .Values.svcConf.mongo.username }}"
{{- if .Values.svcConf.mongo.password }}
- name: MONGO_PASSWORD
  value: "{{ .Values.svcConf.mongo.password }}"
{{- else }}
{{- include "bcs-common.bcspwd.mongodb" ( dict "root" . "envName" "MONGO_PASSWORD" ) }}
{{- end }}
{{- end }}

{{- define "bcs-platform-manager.confCMName" -}}
{{- printf "%s-%s"  (include "bcs-platform-manager.fullname" .) "conf" }}
{{- end }}

{{- define "bcs-platform-manager.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "bcs-platform-manager.fullname" .) .Values.serviceAccount.name }}
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
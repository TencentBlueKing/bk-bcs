{{- define "custom.hookTmplMetadata" -}}
metadata:
  name: {{ .metadata.name }}
  namespace: {{ .metadata.namespace }}
  # HookTemplate 不允许用户自己编辑标签，所以只会有删除保护策略的标签
  {{- if eq .spec.deletionProtectPolicy "Always" }}
  labels:
    io.tencent.bcs.dev/deletion-allow: Always
  {{- end }}
  {{- if .metadata.annotations }}
  annotations:
    {{- include "common.kvSlice2Map" .metadata.annotations | indent 4 }}
  {{- end }}
  {{- if .metadata.resVersion }}
  resourceVersion: {{ .metadata.resVersion | quote }}
  {{- end }}
{{- end }}

{{- define "custom.hookTmplArgs" -}}
{{- if .args }}
args:
  {{- range .args }}
  - name: {{ .key | quote }}
    value: {{ .value | default "" | quote }}
  {{- end }}
{{- end }}
{{- end }}

{{- define "custom.hookTmplMetrics" -}}
{{- if .metrics }}
metrics:
  {{- range .metrics }}
  - name: {{ .name | quote }}
    interval: {{ .interval }}s
    count: {{ .count | default 0 }}
    {{- if .successCondition }}
    successCondition: {{ .successCondition | quote }}
    {{- end }}
    {{- if eq .successPolicy "successfulLimit" }}
    successfulLimit: {{ .successCnt | default 1 }}
    {{- else }}
    consecutiveSuccessfulLimit: {{ .successCnt | default 1 }}
    {{- end }}
    {{- include "custom.hookTmplMetricProvider" . | nindent 4 }}
  {{- end }}
{{- end }}
{{- end }}

{{- define "custom.hookTmplMetricProvider" -}}
provider:
  {{- if eq .hookType "web" }}
  web:
    url: {{ .url | quote }}
    jsonPath: {{ .jsonPath | quote }}
    timeoutSeconds: {{ .timeoutSecs | default 0 }}
  {{- else if eq .hookType "prometheus" }}
  prometheus:
    query: {{ .query | quote }}
    address: {{ .address | quote }}
  {{- else if eq .hookType "kubernetes" }}
  kubernetes:
    function: {{ .function }}
    {{- if .fields }}
    fields:
    {{- range .fields }}
      - path: {{ .key | quote }}
        value: {{ .value | quote }}
    {{- end }}
    {{- end }}
  {{- else }}
  {}
  {{- end }}
{{- end }}

{{- define "custom.gWorkloadMetadata" }}
metadata:
  name: {{ .metadata.name }}
  namespace: {{ .metadata.namespace }}
  labels:
    {{- range .metadata.labels }}
    # 特殊的 LabelKey 单独更新
    {{- if ne .key "io.tencent.bcs.dev/deletion-allow" }}
    {{ .key | quote }}: {{ .value | default "" | quote }}
    {{- end }}
    {{- end }}
    {{- if eq .spec.deletionProtect.policy "Cascading" }}
    # 实例数量为 0 时候才可以删除
    io.tencent.bcs.dev/deletion-allow: Cascading
    {{- else if eq .spec.deletionProtect.policy "Always" }}
    # 任意时候都可以删除，如果没有这个 label key 则不能删除
    io.tencent.bcs.dev/deletion-allow: Always
    {{- end }}
  {{- if .metadata.annotations }}
  annotations:
    {{- include "common.kvSlice2Map" .metadata.annotations | indent 4 }}
  {{- end }}
  {{- if .metadata.resVersion }}
  resourceVersion: {{ .metadata.resVersion | quote }}
  {{- end }}
{{- end }}

{{- define "custom.gworkloadUpdateHook" -}}
hook:
  templateName: {{ .tmplName }}
  {{- if .args }}
  args:
    {{- range .args }}
    - name: {{ .key | quote }}
      value: {{ .value | quote }}
    {{- end }}
  {{- end }}
{{- end }}

{{- define "custom.gworkloadCommonSpec" -}}
selector:
  matchLabels:
    {{- include "common.labelSlice2Map" .metadata.labels | indent 4 }}
replicas: {{ .spec.replicas.cnt | default 0 }}
updateStrategy:
  type: {{ .spec.replicas.updateStrategy }}
  {{- if eq .metadata.kind "GameDeployment" }}
  {{- include "custom.gworkloadUpdateArgs" . | nindent 2 }}
  # 1.27.4+ gameStatefulSet 新增校验规则：如果更新类型为 OnDelete，则不能包含 rollingUpdate 配置
  {{- else if and (eq .metadata.kind "GameStatefulSet") (ne .spec.replicas.updateStrategy "OnDelete") }}
  rollingUpdate:
    {{- include "custom.gworkloadUpdateArgs" . | nindent 4 }}
  {{- end }}
  {{- if eq .spec.replicas.updateStrategy "InplaceUpdate" }}
  inPlaceUpdateStrategy:
    gracePeriodSeconds: {{ .spec.replicas.gracePeriodSecs | default 0 }}
  {{- end }}
{{- if .spec.gracefulManage.preDeleteHook.enabled }}
preDeleteUpdateStrategy:
  {{- include "custom.gworkloadUpdateHook" .spec.gracefulManage.preDeleteHook | nindent 2 }}
{{- end }}
{{- if .spec.gracefulManage.preInplaceHook.enabled }}
preInplaceUpdateStrategy:
  {{- include "custom.gworkloadUpdateHook" .spec.gracefulManage.preInplaceHook | nindent 2 }}
{{- end }}
{{- if .spec.gracefulManage.postInplaceHook.enabled }}
postInplaceUpdateStrategy:
  {{- include "custom.gworkloadUpdateHook" .spec.gracefulManage.postInplaceHook | nindent 2 }}
{{- end }}
{{- if .spec.gracefulManage.preInplaceHook }}
{{- end }}
{{- if .spec.gracefulManage.postInplaceHook }}
{{- end }}
{{- end }}

{{- define "custom.gworkloadUpdateArgs" -}}
partition: {{ .spec.replicas.partition | default 0 }}
maxUnavailable: {{ .spec.replicas.maxUnavailable | default 0 }}{{ if and (.spec.replicas.maxUnavailable) (eq .spec.replicas.muaUnit "percent") }}% {{ end }}
maxSurge: {{ .spec.replicas.maxSurge | default 0 }}{{ if and (.spec.replicas.maxSurge) (eq .spec.replicas.msUnit "percent") }}% {{ end }}
{{- end }}

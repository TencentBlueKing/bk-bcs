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
{{- end }}

{{- define "custom.gdeployUpdateHook" -}}
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

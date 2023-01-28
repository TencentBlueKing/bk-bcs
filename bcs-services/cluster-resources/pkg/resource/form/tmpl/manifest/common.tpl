{{- define "common.metadata" -}}
metadata:
  name: {{ .name }}
  {{- if isNSRequired .kind }}
  namespace: {{ .namespace }}
  {{- end }}
  {{- if .labels }}
  labels:
    {{- include "common.labelSlice2Map" .labels | indent 4 }}
  {{- end }}
  {{- if .annotations }}
  annotations:
    {{- include "common.kvSlice2Map" .annotations | indent 4 }}
  {{- end }}
  {{- if and (canRenderResVersion .kind) .resVersion }}
  resourceVersion: {{ .resVersion | quote }}
  {{- end }}
{{- end }}

{{- define "common.labelSlice2Map" -}}
{{- range . }}
# 跳过一些特殊的键，这些键需要走独立的 labels 渲染逻辑，比如 custom.gdeployMetadata
{{- if ne .key "io.tencent.bcs.dev/deletion-allow" }}
{{ .key | quote }}: {{ .value | default "" | quote }}
{{- end }}
{{- else }}
{}
{{- end }}
{{- end }}

{{- define "common.kvSlice2Map" -}}
{{- range . }}
{{ .key | quote }}: {{ .value | default "" | quote }}
{{- else }}
{}
{{- end }}
{{- end }}

{{- define "common.splitStr2Slice" -}}
{{- range $_, $item := splitList `,` . }}
- {{ $item | quote }}
{{- else }}
[]
{{- end }}
{{- end }}

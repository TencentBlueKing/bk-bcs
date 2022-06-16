{{- define "common.metadata" -}}
metadata:
  name: {{ .name }}
  namespace: {{ .namespace }}
  {{- if .labels }}
  labels:
    {{- include "common.kvSlice2Map" .labels | indent 4 }}
  {{- end }}
  {{- if .annotations }}
  annotations:
    {{- include "common.kvSlice2Map" .annotations | indent 4 }}
  {{- end }}
{{- end }}

{{- define "common.kvSlice2Map" -}}
{{- range . }}
{{ .key | quote }}: {{ .value | quote }}
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

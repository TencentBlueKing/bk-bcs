{{- define "common.metadata" -}}
metadata:
  name: {{ .name }}
  namespace: {{ .namespace }}
  labels:
    {{- range .labels }}
    {{ .key | quote }}: {{ .value | quote }}
    {{- else }}
    {}
    {{- end }}
  annotations:
    {{- range .annotations }}
    {{ .key | quote }}: {{ .value | quote }}
    {{- else }}
    {}
    {{- end }}
{{- end }}
{{- define "nodeAffinity.required" -}}
requiredDuringSchedulingIgnoredDuringExecution:
  nodeSelectorTerms:
    {{- range . }}
    {{- if eq .priority "required" }}
    {{- if or .selector.expressions .selector.fields }}
    - {{ if .selector.expressions -}}
      matchExpressions:
      {{- range .selector.expressions }}
      - key: {{ .key | quote }}
        operator: {{ .op }}
        values:
          {{- include "common.splitStr2Slice" .values | indent 10 }}
      {{- end }}
      {{- end }}
      {{- if .selector.fields }}
      matchFields:
      {{- range .selector.fields }}
      - key: {{ .key | quote }}
        operator: {{ .op }}
        values:
          {{- include "common.splitStr2Slice" .values | indent 10 }}
      {{- end }}
      {{- end }}
    {{- end }}
    {{- end }}
    {{- end }}
{{- end }}

{{- define "nodeAffinity.preferred" -}}
preferredDuringSchedulingIgnoredDuringExecution:
  {{- range . }}
  {{- if eq .priority "preferred" }}
  - weight: {{ .weight }}
    {{- if or .selector.expressions .selector.fields }}
    preference:
      {{- if .selector.expressions }}
      matchExpressions:
      {{- range .selector.expressions }}
      - key: {{ .key | quote }}
        operator: {{ .op }}
        values:
          {{- include "common.splitStr2Slice" .values | indent 10 }}
      {{- end }}
      {{- end }}
      {{- if .selector.fields }}
      matchFields:
      {{- range .selector.fields }}
      - key: {{ .key | quote }}
        operator: {{ .op }}
        values:
          {{- include "common.splitStr2Slice" .values | indent 10 }}
      {{- end }}
      {{- end }}
    {{- end }}
  {{- end }}
  {{- end }}
{{- end }}

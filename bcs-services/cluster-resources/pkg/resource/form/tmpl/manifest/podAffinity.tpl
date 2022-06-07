{{- define "podAffinity.required" -}}
requiredDuringSchedulingIgnoredDuringExecution:
  {{- range . }}
  {{- if eq .priority "required" }}
  - topologyKey: {{ .topologyKey | quote }}
    {{- if .namespaces }}
    namespaces:
      {{- toYaml .namespaces | nindent 6 }}
    {{- end }}
    {{- if or .selector.expressions .selector.labels }}
    labelSelector:
      {{- if .selector.expressions }}
      matchExpressions:
        {{- range .selector.expressions }}
        - key: {{ .key | quote }}
          operator: {{ .op }}
          values:
            {{- include "common.splitStr2Slice" .values | indent 12 }}
        {{- end }}
      {{- end }}
      {{- if .selector.labels }}
      matchLabels:
        {{- include "common.kvSlice2Map" .selector.labels | indent 8 }}
      {{- end }}
    {{- end }}
  {{- end }}
  {{- end }}
{{- end }}

{{- define "podAffinity.preferred" -}}
preferredDuringSchedulingIgnoredDuringExecution:
  {{- range . }}
  {{- if eq .priority "preferred" }}
  - weight: {{ .weight }}
    podAffinityTerm:
      topologyKey: {{ .topologyKey | quote }}
      {{- if .namespaces }}
      namespaces:
        {{- toYaml .namespaces | nindent 8 }}
      {{- end }}
      {{- if or .selector.expressions .selector.labels }}
      labelSelector:
        {{- if .selector.expressions }}
        matchExpressions:
          {{- range .selector.expressions }}
          - key: {{ .key | quote }}
            operator: {{ .op }}
            values:
              {{- include "common.splitStr2Slice" .values | indent 14 }}
          {{- end }}
        {{- end }}
        {{- if .selector.labels }}
        matchLabels:
          {{- include "common.kvSlice2Map" .selector.labels | indent 10 }}
        {{- end }}
      {{- end }}
  {{- end }}
  {{- end }}
{{- end }}

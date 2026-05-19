{{/*
Common labels
*/}}
{{- define "nginx-drplan.labels" -}}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/part-of: nginx-drplan
{{- end }}

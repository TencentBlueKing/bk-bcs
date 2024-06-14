{{- define "network.ingress.backend" }}
{{- if or (eq .apiVersion "extensions/v1beta1") (eq .apiVersion "networking.k8s.io/v1beta1") -}}
serviceName: {{ .svcName }}
servicePort: {{ .svcPort }}
{{- else -}}
service:
  name: {{ .svcName }}
  port:
    {{- if typeIs "string" .svcPort }}
    name: {{ .svcPort }}
    {{- else }}
    number: {{ .svcPort | int }}
    {{- end }}
{{- end }}
{{- end }}

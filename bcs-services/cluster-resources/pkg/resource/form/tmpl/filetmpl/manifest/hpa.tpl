{{- define "hpa.refObj" -}}
scaleTargetRef:
  kind: {{ .kind }}
  apiVersion: {{ .apiVersion }}
  name: {{ .resName }}
minReplicas: {{ .minReplicas }}
maxReplicas: {{ .maxReplicas }}
{{- end }}

{{- define "hpa.resMetric" -}}
{{- range .items }}
- type: Resource
  resource:
    name: {{ .name | quote }}
    {{- include "hpa.resMetricTarget" . | nindent 4 }}
{{- end }}
{{- end }}

{{- define "hpa.containerResMetric" -}}
{{- range .items }}
- type: ContainerResource
  containerResource:
    name: {{ .name | quote }}
    container: {{ .containerName | quote }}
    {{- include "hpa.metricTarget" . | nindent 4 }}
{{- end }}
{{- end }}

{{- define "hpa.externalMetric" -}}
{{- range .items }}
- type: External
  external:
    metric:
      name: {{ .name | quote }}
      {{- include "hpa.metricSelector" .selector | indent 6 }}
    {{- include "hpa.metricTarget" . | nindent 4 }}
{{- end }}
{{- end }}

{{- define "hpa.objMetric" -}}
{{- range .items }}
- type: Object
  object:
    describedObject:
      apiVersion: {{ .apiVersion }}
      kind: {{ .kind }}
      name: {{ .resName | quote }}
    metric:
      name: {{ .name | quote }}
      {{- include "hpa.metricSelector" .selector | indent 6 }}
    {{- include "hpa.metricTarget" . | nindent 4 }}
{{- end }}
{{- end }}

{{- define "hpa.podMetric" -}}
{{- range .items }}
- type: Pods
  pods:
    metric:
      name: {{ .name | quote }}
      {{- include "hpa.metricSelector" .selector | indent 6 }}
    {{- include "hpa.metricTarget" . | nindent 4 }}
{{- end }}
{{- end }}

{{- define "hpa.metricSelector" -}}
{{- if . }}
selector:
  {{- if .expressions }}
  matchExpressions:
  {{- range .expressions }}
  - key: {{ .key | quote }}
    operator: {{ .op }}
    {{- if .values }}
    values:
      {{- include "common.splitStr2Slice" .values | indent 6 }}
    {{- end }}
  {{- end }}
  {{- end }}
  {{- if .labels }}
  matchLabels:
    {{- include "common.kvSlice2Map" .labels | indent 6 }}
  {{- end }}
{{- end }}
{{- end }}

# Resource 指标特有子模板，为 CPU, Memory 指标类型做特化
{{- define "hpa.resMetricTarget" -}}
target:
  type: {{ .type }}
  {{- if eq .type "AverageValue" }}
  {{- if eq .name "cpu" }}
  averageValue: {{ .cpuVal }}m
  {{- end }}
  {{- if eq .name "memory" }}
  averageValue: {{ .memVal }}Mi
  {{- end }}
  {{- end }}
  {{- if eq .type "Utilization" }}
  averageUtilization: {{ .percent }}
  {{- end }}
{{- end }}

{{- define "hpa.metricTarget" -}}
target:
  type: {{ .type }}
  {{- if eq .type "AverageValue" }}
  averageValue: {{ .value }}
  {{- end }}
  {{- if eq .type "Value" }}
  value: {{ .value }}
  {{- end }}
  {{- if eq .type "Utilization" }}
  averageUtilization: {{ .value }}
  {{- end }}
{{- end }}

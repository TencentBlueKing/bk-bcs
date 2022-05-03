{{- define "workload.podTemplate" -}}
template:
  metadata:
    labels:
      {{- include "common.kvSlice2Map" .metadata.labels | indent 6 }}
  spec:
    {{- include "container.containers" .containerGroup.containers | nindent 4 }}
    {{- include "container.initContainers" .containerGroup.initContainers | nindent 4 }}
    # affinity
    {{- if .spec.affinity }}
    affinity:
      {{- include "workload.affinity" .spec.affinity | indent 6 }}
    {{- end }}
    # toleration
    {{- if .spec.toleration }}
    tolerations:
      {{- include "workload.toleration" .spec.toleration | indent 6 }}
    {{- end }}
    # nodeSelect
    {{- include "workload.nodeSelect" .spec.nodeSelect | indent 4 }}
    # networking
    {{- include "workload.network" .spec.networking | nindent 4 }}
    # security
    {{- if .spec.security }}
    securityContext:
      {{- include "workload.security" .spec.security | indent 6 }}
    {{- end }}
    # other
    {{- include "workload.specOther" .spec.other | indent 4 }}
    {{- if .volume }}
    volumes:
      {{- include "workload.volume" .volume | indent 6 }}
    {{- end }}
{{- end }}

{{- define "workload.affinity" -}}
{{- $podAffinity := filterMatchKVFormSlice .podAffinity "type" "affinity" }}
{{- if $podAffinity }}
podAffinity:
  {{- if matchKVInSlice $podAffinity "priority" "required" }}
  {{- include "podAffinity.required" $podAffinity | nindent 2 }}
  {{- end }}
  {{- if matchKVInSlice $podAffinity "priority" "preferred" }}
  {{- include "podAffinity.preferred" $podAffinity | nindent 2 }}
  {{- end }}
{{- end }}
{{- $podAntiAffinity := filterMatchKVFormSlice .spec.affinity.podAffinity "type" "antiAffinity" }}
{{- if $podAntiAffinity }}
podAntiAffinity:
  {{- if matchKVInSlice $podAntiAffinity "priority" "required" }}
  {{- include "podAffinity.required" $podAntiAffinity | nindent 2 }}
  {{- end }}
  {{- if matchKVInSlice $podAffinity "priority" "preferred" }}
  {{- include "podAffinity.preferred" $podAntiAffinity | nindent 2 }}
  {{- end }}
{{- end }}
{{- if .nodeAffinity }}
nodeAffinity:
  {{- if matchKVInSlice .nodeAffinity "priority" "required" }}
  {{- include "nodeAffinity.required" .nodeAffinity | nindent 2 }}
  {{- end }}
  {{- if matchKVInSlice .nodeAffinity "priority" "preferred" }}
  {{- include "nodeAffinity.preferred" .nodeAffinity | nindent 2 }}
  {{- end }}
{{- end }}
{{- end }}

{{- define "workload.toleration" -}}
{{- range .rules }}
- key: {{ .key | quote }}
  operator: {{ .op }}
  effect: {{ .effect }}
  {{- if .value }}
  value: {{ .value | quote }}
  {{- end }}
  {{- if .tolerationSecs }}
  tolerationSeconds: {{ .tolerationSecs }}
  {{- end }}
{{- else }}
[]
{{- end }}
{{- end }}

{{- define "workload.nodeSelect" -}}
{{- if eq .type "specificNode" }}
nodeName: {{ .nodeName }}
{{- else if eq .type "schedulingRule" }}
nodeSelector:
  {{- include "common.kvSlice2Map" .selector | indent 2 }}
{{- end }}
{{- end }}

{{- define "workload.network" -}}
dnsPolicy: {{ .dnsPolicy }}
{{- if .hostIPC }}
hostIPC: {{ .hostIPC }}
{{- end }}
{{- if .hostNetwork }}
hostNetwork: {{ .hostNetwork }}
{{- end }}
{{- if .hostPID }}
hostPID: {{ .hostPID }}
{{- end }}
{{- if .shareProcessNamespace }}
shareProcessNamespace: {{ .shareProcessNamespace }}
{{- end }}
{{- if .hostname }}
hostname: {{ .hostname }}
{{- end }}
{{- if .subdomain }}
subdomain: {{ .subdomain }}
{{- end }}
{{- if or .nameServers .searches .dnsResolverOpts }}
dnsConfig:
  {{- if .nameServers }}
  nameservers:
    {{- toYaml .nameServers | nindent 4 }}
  {{- end }}
  {{- if .searches }}
  searches:
    {{- toYaml .searches | nindent 4 }}
  {{- end }}
  {{- if .dnsResolverOpts }}
  options:
    {{- range .dnsResolverOpts }}
    - name: {{ .name | quote }}
      value: {{ .value | quote }}
    {{- end }}
  {{- end }}
{{- end }}
{{- if .hostAliases }}
hostAliases:
  {{- range .hostAliases }}
  - ip: {{ .ip | quote }}
    hostnames:
      {{- include "common.splitStr2Slice" .alias | indent 6 }}
  {{- end }}
{{- end }}
{{- end }}

{{- define "workload.security" -}}
{{- if .runAsUser }}
runAsUser: {{ .runAsUser }}
{{- end }}
{{- if .runAsNonRoot }}
runAsNonRoot: {{ .runAsNonRoot }}
{{- end }}
{{- if .runAsGroup }}
runAsGroup: {{ .runAsGroup }}
{{- end }}
{{- if .fsGroup }}
fsGroup: {{ .fsGroup }}
{{- end }}
{{- if .seLinuxOpt }}
seLinuxOptions:
  {{- range $k, $v := .seLinuxOpt }}
  {{ $k | quote }}: {{ $v | quote }}
  {{- end }}
{{- end }}
{{- end }}

{{- define "workload.specOther" -}}
{{- if .restartPolicy }}
restartPolicy: {{ .restartPolicy }}
{{- end }}
{{- if .terminationGracePeriodSecs }}
terminationGracePeriodSeconds: {{ .terminationGracePeriodSecs }}
{{- end }}
imagePullSecrets:
  {{- range .imagePullSecrets }}
  - name: {{ . | quote }}
  {{- else }}
  []
  {{- end }}
{{- if .saName }}
serviceAccountName: {{ .saName }}
{{- end }}
{{- end }}

{{- define "workload.volume" -}}
{{- range .pvc }}
- name: {{ .name }}
  persistentVolumeClaim:
    claimName: {{ .pvcName }}
    {{- if .readOnly }}
    readOnly: {{ .readOnly }}
    {{- end }}
{{- end }}
{{- range .hostPath }}
- name: {{ .name }}
  hostPath:
    path: {{ .path | quote }}
    type: {{ .type }}
{{- end }}
{{- range .configMap }}
- name: {{ .name }}
  configMap:
    defaultMode: {{ .defaultMode }}
    name: {{ .cmName }}
    {{- if .items }}
    items:
      {{- range .items }}
      - key: {{ .key | quote }}
        path: {{ .path | quote }}
      {{- end }}
    {{- end }}
{{- end }}
{{- range .secret }}
- name: {{ .name }}
  secret:
    defaultMode: {{ .defaultMode }}
    secretName: {{ .secretName }}
    {{- if .items }}
    items:
      {{- range .items }}
      - key: {{ .key | quote }}
        path: {{ .path | quote }}
      {{- end }}
    {{- end }}
{{- end }}
{{- range .emptyDir }}
- name: {{ .name }}
  emptyDir: {}
{{- end }}
{{- range .nfs }}
- name: {{ .name }}
  nfs:
    path: {{ .path | quote }}
    server: {{ .server | quote }}
    {{- if .readOnly }}
    readOnly: {{ .readOnly }}
    {{- end }}
{{- end }}
{{- end }}

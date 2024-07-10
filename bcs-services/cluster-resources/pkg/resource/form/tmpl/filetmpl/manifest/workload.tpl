{{- define "workload.podTemplate" -}}
template:
  {{- if .spec.labels.templateLabels }}
  metadata:
    labels:
      {{- include "common.labelSlice2Map" .spec.labels.templateLabels | indent 6 }}
  {{- end }}
  {{- include "workload.podSpec" . | nindent 2 }}
{{- end }}

{{- define "workload.podSpec" -}}
spec:
  {{- include "container.containers" .containerGroup.containers | nindent 2 }}
  {{- include "container.initContainers" .containerGroup.initContainers | nindent 2 }}
  # affinity
  {{- if .spec.affinity }}
  affinity:
    {{- include "workload.affinity" .spec.affinity | indent 4 }}
  {{- end }}
  # toleration
  {{- if .spec.toleration }}
  tolerations:
    {{- include "workload.toleration" .spec.toleration | indent 4 }}
  {{- end }}
  # nodeSelect
  {{- include "workload.nodeSelect" .spec.nodeSelect | indent 2 }}
  # networking
  {{- include "workload.network" .spec.networking | nindent 2 }}
  # security
  {{- if .spec.security }}
  securityContext:
    {{- include "workload.security" .spec.security | indent 4 }}
  {{- end }}
  # readinessGates
  {{- include "workload.readinessGates" .spec.readinessGates | indent 2 }}
  # other
  {{- include "workload.specOther" .spec.other | indent 2 }}
  {{- if .volume }}
  volumes:
    {{- include "workload.volume" .volume | indent 4 }}
  {{- end }}
{{- end }}

{{- define "workload.stsVolumeClaimTmpl" -}}
{{- if .spec.volumeClaimTmpl.claims }}
volumeClaimTemplates:
  {{- range .spec.volumeClaimTmpl.claims }}
  - metadata:
      name: {{ .pvcName }}
    spec:
      accessModes:
        {{- range .accessModes }}
        - {{ . }}
        {{- else }}
        []
        {{- end }}
      {{- if eq .claimType "useExistPV" }}
      volumeName: {{ .pvName }}
      {{- else if eq .claimType "createBySC" }}
      storageClassName: {{ .scName }}
      resources:
        requests:
          storage: {{ .storageSize }}Gi
      {{- end }}
  {{- end }}
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
{{- $podAntiAffinity := filterMatchKVFormSlice .podAffinity "type" "antiAffinity" }}
{{- if $podAntiAffinity }}
podAntiAffinity:
  {{- if matchKVInSlice $podAntiAffinity "priority" "required" }}
  {{- include "podAffinity.required" $podAntiAffinity | nindent 2 }}
  {{- end }}
  {{- if matchKVInSlice $podAntiAffinity "priority" "preferred" }}
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
{{- if .dnsPolicy }}
dnsPolicy: {{ .dnsPolicy }}
{{- end }}
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
hostname: {{ .hostname | quote }}
{{- end }}
{{- if .subdomain }}
subdomain: {{ .subdomain | quote }}
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
    {{- if .alias }}
    hostnames:
      {{- include "common.splitStr2Slice" .alias | indent 6 }}
    {{- end }}
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

{{- define "workload.readinessGates" -}}
{{- if .readinessGates }}
readinessGates:
  {{- range .readinessGates }}
  - conditionType: {{ . | quote }}
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
{{- if .imagePullSecrets }}
imagePullSecrets:
  {{- range .imagePullSecrets }}
  - name: {{ . | quote }}
  {{- else }}
  []
  {{- end }}
{{- end }}
{{- if .saName }}
serviceAccountName: {{ .saName }}
{{- end }}
{{- end }}

{{- define "workload.volume" -}}
{{- range .pvc }}
- name: {{ .name | quote }}
  persistentVolumeClaim:
    claimName: {{ .pvcName }}
    {{- if .readOnly }}
    readOnly: {{ .readOnly }}
    {{- end }}
{{- end }}
{{- range .hostPath }}
- name: {{ .name | quote }}
  hostPath:
    path: {{ .path | quote }}
    type: {{ .type }}
{{- end }}
{{- range .configMap }}
- name: {{ .name | quote }}
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
- name: {{ .name | quote }}
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
- name: {{ .name | quote }}
  emptyDir: {}
{{- end }}
{{- range .nfs }}
- name: {{ .name | quote }}
  nfs:
    path: {{ .path | quote }}
    server: {{ .server | quote }}
    {{- if .readOnly }}
    readOnly: {{ .readOnly }}
    {{- end }}
{{- end }}
{{- end }}

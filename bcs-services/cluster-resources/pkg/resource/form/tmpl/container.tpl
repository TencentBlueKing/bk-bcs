{{- define "container.containers" -}}
containers:
  {{- range . }}
  - name: {{ .basic.name }}
    image: {{ .basic.image }}
    {{- if .basic.pullPolicy }}
    imagePullPolicy: {{ .basic.pullPolicy }}
    {{- end }}
    {{- include "container.command" .command | indent 4 }}
    {{- include "container.service" .service | nindent 4 }}
    {{- include "container.envs" .envs | indent 4 }}
    {{- if .healthz.readinessProbe }}
    readinessProbe:
      {{- include "container.probe" .healthz.readinessProbe | indent 6 }}
    {{- end }}
    {{- if .healthz.livenessProbe }}
    livenessProbe:
      {{- include "container.probe" .healthz.livenessProbe | indent 6 }}
    {{- end }}
    {{- if .resource }}
    {{- include "container.resource" .resource | nindent 4 }}
    {{- end }}
    {{- if .security }}
    {{- include "container.security" .security | nindent 4 }}
    {{- end }}
    {{- if .mount }}
    {{- include "container.mount" .mount | nindent 4 }}
    {{- end }}
  {{- else }}
  []
  {{- end }}
{{- end }}

{{- define "container.initContainers" -}}
initContainers:
  {{- range . }}
  - name: {{ .basic.name }}
    image: {{ .basic.image }}
    {{- if .basic.pullPolicy }}
    imagePullPolicy: {{ .basic.pullPolicy }}
    {{- end }}
    {{- include "container.command" .command | indent 4 }}
    {{- include "container.envs" .envs | indent 4 }}
    {{- if .resource }}
    {{- include "container.resource" .resource | nindent 4 }}
    {{- end }}
    {{- if .security }}
    {{- include "container.security" .security | nindent 4 }}
    {{- end }}
  {{- else }}
  []
  {{- end }}
{{- end }}

{{- define "container.command" -}}
{{- if .workingDir }}
workingDir: {{ .workingDir }}
{{- end }}
{{- if .stdin }}
stdin: {{ .stdin }}
{{- end }}
{{- if .stdinOnce }}
stdinOnce: {{ .stdinOnce }}
{{- end }}
{{- if .tty }}
tty: {{ .tty }}
{{- end }}
{{- if .command }}
command:
  {{- toYaml .command | nindent 2 }}
{{- end }}
{{- if .args }}
args:
  {{- toYaml .args | nindent 2 }}
{{- end }}
{{- end }}

{{- define "container.service" -}}
ports:
  {{- range .ports }}
  - containerPort: {{ .containerPort }}
    {{- if .name }}
    name: {{ .name }}
    {{- end }}
    {{- if .protocol }}
    protocol: {{ .protocol }}
    {{- end }}
    {{- if .hostport }}
    hostPort: {{ .hostPort }}
    {{- end }}
  {{- else }}
  []
  {{- end }}
{{- end }}

{{- define "container.envs" -}}
{{- if or (matchKVInSlice .vars "type" "keyValue") (matchKVInSlice .vars "type" "podField") (matchKVInSlice .vars "type" "resource") (matchKVInSlice .vars "type" "configMapKey") (matchKVInSlice .vars "type" "secretKey") }}
envs:
  {{- range .vars }}
  {{- if eq .type "keyValue" }}
  - name: {{ .name }}
    value: {{ .value }}
  {{- else if eq .type "podField" }}
  - name: {{ .name }}
    valueFrom:
      fieldRef:
        apiVersion: v1
        fieldPath: {{ .value }}
  {{- else if eq .type "resource" }}
  - name: {{ .name }}
    valueForm:
      resourceFieldRef:
        containerName: {{ .source }}
        divisor: 0
        resource: {{ .value }}
  {{- else if eq .type "configMapKey" }}
  - name: {{ .name }}
    valueForm:
      configMapKeyRef:
        name: {{ .source }}
        key: {{ .value }}
  {{- else if eq .type "secretKey" }}
  - name: {{ .name }}
    valueForm:
      secretKeyRef:
        name: {{ .source }}
        key: {{ .value }}
  {{- end }}
  {{- end }}
{{- end }}
{{- if or (matchKVInSlice .vars "type" "configMap") (matchKVInSlice .vars "type" "secret") }}
envForm:
  {{- range .vars }}
  {{- if eq .type "configMap" }}
  - prefix: {{ .name }}
    configMapRef:
      name: {{ .source }}
  {{- else if eq .type "secret" }}
  - prefix: {{ .name }}
    secretRef:
      name: {{ .source }}
  {{- end }}
  {{- end }}
{{- end }}
{{- end }}

{{- define "container.probe" -}}
{{- if .periodSecs }}
periodSeconds: {{ .periodSecs }}
{{- end }}
{{- if .initialDelaySecs }}
initialDelaySeconds: {{ .initialDelaySecs }}
{{- end }}
{{- if .timeoutSecs }}
timeoutSeconds: {{ .timeoutSecs }}
{{- end }}
{{- if .successThreshold }}
successThreshold: {{ .successThreshold }}
{{- end }}
{{- if .failureThreshold }}
failureThreshold: {{ .failureThreshold }}
{{- end }}
{{- if eq .type "httpGet" }}
httpGet:
  scheme: HTTP
  path: {{ .path }}
  port: {{ .port }}
{{- else if eq .type "tcpSocket" }}
tcpSocket:
  port: {{ .port }}
{{- else if eq .type "exec" }}
exec:
  command:
    {{- toYaml (default list .command) | nindent 4 }}
{{- end }}
{{- end }}

{{- define "container.resource" -}}
resources:
  {{- if .requests }}
  requests:
    {{- if .requests.cpu }}
    cpu: {{ printf "%.0fm" .requests.cpu }}
    {{- end }}
    {{- if .requests.memory }}
    memory: {{ printf "%.0fMi" .requests.memory }}
    {{- end }}
  {{- end }}
  {{- if .limits }}
  limits:
    {{- if .limits.cpu }}
    cpu: {{ printf "%.0fm" .limits.cpu }}
    {{- end }}
    {{- if .limits.memory }}
    memory: {{ printf "%.0fMi" .limits.memory }}
    {{- end }}
  {{- end }}
{{- end }}

{{- define "container.security" -}}
securityContext:
  {{- if .privileged }}
  privileged: {{ .privileged }}
  {{- end }}
  {{- if .allowPrivilegeEscalation }}
  allowPrivilegeEscalation: {{ .allowPrivilegeEscalation }}
  {{- end }}
  {{- if .runAsNonRoot }}
  runAsNonRoot: {{ .runAsNonRoot }}
  {{- end }}
  {{- if .readOnlyRootFilesystem }}
  readOnlyRootFilesystem: {{ .readOnlyRootFilesystem }}
  {{- end }}
  {{- if .runAsUser }}
  runAsUser: {{ .runAsUser }}
  {{- end }}
  {{- if .runAsGroup }}
  runAsGroup: {{ .runAsGroup }}
  {{- end }}
  {{- if .procMount }}
  procMount: {{ .procMount | quote }}
  {{- end }}
  {{- if or .capabilities.add .capabilities.drop }}
  capabilities:
    {{- if .capabilities.add }}
    add:
      {{- toYaml .capabilities.add | nindent 6 }}
    {{- end }}
    {{- if .capabilities.drop }}
    drop:
      {{- toYaml .capabilities.drop | nindent 6 }}
    {{- end }}
  {{- end }}
  {{- if .seLinuxOpt }}
  seLinuxOptions:
    {{- range $k, $v := .seLinuxOpt }}
    {{ $k | quote }}: {{ $v | quote }}
    {{- end }}
  {{- end }}
{{- end }}

{{- define "container.mount" -}}
{{- if .volumes }}
volumeMounts:
  {{- range .volumes }}
  - name: {{ .name }}
    {{- if .mountPath }}
    mountPath: {{ .mountPath }}
    {{- end }}
    {{- if .subPath }}
    subPath: {{ .subPath }}
    {{- end }}
    {{- if .readOnly }}
    readOnly: {{ .readOnly }}
    {{- end }}
  {{- end }}
{{- end }}
{{- end }}

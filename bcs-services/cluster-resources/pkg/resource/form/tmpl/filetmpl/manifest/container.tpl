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
    {{- if and .healthz.readinessProbe .healthz.readinessProbe.enabled }}
    readinessProbe:
      {{- include "container.probe" .healthz.readinessProbe | indent 6 }}
    {{- end }}
    {{- if and .healthz.livenessProbe .healthz.livenessProbe.enabled }}
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
{{- if . }}
initContainers:
  {{- range . }}
  - name: {{ .basic.name }}
    image: {{ .basic.image }}
    {{- if .basic.pullPolicy }}
    imagePullPolicy: {{ .basic.pullPolicy }}
    {{- end }}
    {{- include "container.command" .command | indent 4 }}
    {{- include "container.service" .service | nindent 4 }}
    {{- include "container.envs" .envs | indent 4 }}
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
{{- if .ports }}
ports:
  {{- range .ports }}
  - containerPort: {{ .containerPort }}
    {{- if .name }}
    name: {{ .name }}
    {{- end }}
    {{- if .protocol }}
    protocol: {{ .protocol }}
    {{- end }}
    {{- if .hostPort }}
    hostPort: {{ .hostPort }}
    {{- end }}
  {{- else }}
  []
  {{- end }}
{{- end }}
{{- end }}

{{- define "container.envs" -}}
{{- if or (matchKVInSlice .vars "type" "keyValue") (matchKVInSlice .vars "type" "podField") (matchKVInSlice .vars "type" "resource") (matchKVInSlice .vars "type" "configMapKey") (matchKVInSlice .vars "type" "secretKey") }}
env:
  {{- range .vars }}
  {{- if eq .type "keyValue" }}
  - name: {{ .name | quote }}
    value: {{ .value | quote }}
  {{- else if eq .type "podField" }}
  - name: {{ .name | quote }}
    valueFrom:
      fieldRef:
        apiVersion: v1
        fieldPath: {{ .value | quote }}
  {{- else if eq .type "resource" }}
  - name: {{ .name | quote }}
    valueFrom:
      resourceFieldRef:
        containerName: {{ .source | quote }}
        divisor: 0
        resource: {{ .value | quote }}
  {{- else if eq .type "configMapKey" }}
  - name: {{ .name | quote }}
    valueFrom:
      configMapKeyRef:
        name: {{ .source | quote }}
        key: {{ .value | quote }}
  {{- else if eq .type "secretKey" }}
  - name: {{ .name | quote }}
    valueFrom:
      secretKeyRef:
        name: {{ .source | quote }}
        key: {{ .value | quote }}
  {{- end }}
  {{- end }}
{{- end }}
{{- if or (matchKVInSlice .vars "type" "configMap") (matchKVInSlice .vars "type" "secret") }}
envFrom:
  {{- range .vars }}
  {{- if eq .type "configMap" }}
  - prefix: {{ .name | quote }}
    configMapRef:
      name: {{ .source | quote }}
  {{- else if eq .type "secret" }}
  - prefix: {{ .name | quote }}
    secretRef:
      name: {{ .source | quote }}
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
    cpu: {{ .requests.cpu }}m
    {{- end }}
    {{- if .requests.memory }}
    memory: {{ .requests.memory }}Mi
    {{- end }}
    {{- if index . "requests" "ephemeral-storage" }}
    ephemeral-storage: {{ index . "requests" "ephemeral-storage" }}Gi
    {{- end }}
    {{- if .requests.extra }}
    {{- range .requests.extra }}
    {{ .key }}: {{ .value | quote }}
    {{- end }}
    {{- end }}
  {{- end }}
  {{- if .limits }}
  limits:
    {{- if .limits.cpu }}
    cpu: {{ .limits.cpu }}m
    {{- end }}
    {{- if .limits.memory }}
    memory: {{ .limits.memory }}Mi
    {{- end }}
    {{- if index . "limits" "ephemeral-storage" }}
    ephemeral-storage: {{ index . "limits" "ephemeral-storage" }}Gi
    {{- end }}
    {{- if .limits.extra }}
    {{- range .limits.extra }}
    {{ .key }}: {{ .value | quote }}
    {{- end }}
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

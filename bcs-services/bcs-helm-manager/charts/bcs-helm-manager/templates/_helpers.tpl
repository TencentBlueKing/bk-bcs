{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "bcs-helm-manager.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "bcs-helm-manager.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "bcs-helm-manager.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "bcs-helm-manager.labels" -}}
helm.sh/chart: {{ include "bcs-helm-manager.chart" . }}
{{ include "bcs-helm-manager.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "bcs-helm-manager.selectorLabels" -}}
app.kubernetes.io/name: {{ include "bcs-helm-manager.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "bcs-helm-manager.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "bcs-helm-manager.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create a default fully qualified app name for etcd subchart
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "bcs-helm-manager.etcd.fullname" -}}
{{- if .Values.etcd.fullnameOverride -}}
{{- .Values.etcd.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default "etcd" .Values.etcd.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Return the Etcd Address
*/}}
{{- define "bcs-helm-manager.etcd.address" -}}
{{- if .Values.etcd.enabled }}
    {{- printf "%s" .Values.etcd.address -}}
{{- else -}}
    {{- printf "%s" .Values.externalEtcd.address -}}
{{- end -}}
{{- end -}}

{{/*
Return the Etcd CA
*/}}
{{- define "bcs-helm-manager.etcd.ca" -}}
{{- if .Values.etcd.enabled }}
    {{- printf "%s" .Values.etcd.auth.client.caFilename -}}
{{- else -}}
    {{- printf "%s" .Values.externalEtcd.ca -}}
{{- end -}}
{{- end -}}

{{/*
Return the Etcd Cert
*/}}
{{- define "bcs-helm-manager.etcd.cert" -}}
{{- if .Values.etcd.enabled }}
    {{- printf "%s" .Values.etcd.auth.client.certFilename -}}
{{- else -}}
    {{- printf "%s" .Values.externalEtcd.cert -}}
{{- end -}}
{{- end -}}

{{/*
Return the Etcd Key
*/}}
{{- define "bcs-helm-manager.etcd.key" -}}
{{- if .Values.etcd.enabled }}
    {{- printf "%s" .Values.etcd.auth.client.certKeyFilename -}}
{{- else -}}
    {{- printf "%s" .Values.externalEtcd.key -}}
{{- end -}}
{{- end -}}

{{/*
Create a default fully qualified app name for mongodb subchart
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "bcs-helm-manager.mongodb.fullname" -}}
{{- if .Values.mongodb.fullnameOverride -}}
{{- .Values.mongodb.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default "mongodb" .Values.mongodb.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Return the Mongodb Username
*/}}
{{- define "bcs-helm-manager.mongodb.username" -}}
{{- if .Values.mongodb.enabled }}
    {{- printf "%s" .Values.mongodb.auth.username -}}
{{- else -}}
    {{- printf "%s" .Values.externalMongo.username -}}
{{- end -}}
{{- end -}}

{{/*
Return the Mongodb Password
*/}}
{{- define "bcs-helm-manager.mongodb.password" -}}
{{- if .Values.mongodb.enabled }}
    {{- printf "%s" .Values.mongodb.auth.password -}}
{{- else -}}
    {{- printf "%s" .Values.externalMongo.password -}}
{{- end -}}
{{- end -}}

{{/*
Return the Mongodb Database
*/}}
{{- define "bcs-helm-manager.mongodb.database" -}}
{{- if .Values.mongodb.enabled }}
    {{- printf "%s" .Values.mongodb.auth.database -}}
{{- else -}}
    {{- printf "%s" .Values.externalMongo.database -}}
{{- end -}}
{{- end -}}

{{/*
Return the Mongodb Hostname
*/}}
{{- define "bcs-helm-manager.mongodb.host" -}}
{{- if .Values.mongodb.enabled }}
    {{- printf "%s" (include "bcs-helm-manager.mongodb.fullname" .) -}}
{{- else -}}
    {{- printf "%s" .Values.externalMongo.host -}}
{{- end -}}
{{- end -}}

{{/*
Return the Mongodb Address
*/}}
{{- define "bcs-helm-manager.mongodb.address" -}}
{{- if .Values.mongodb.enabled }}
    {{- printf "%s" .Values.mongodb.address -}}
{{- else -}}
    {{- printf "%s" .Values.externalMongo.address -}}
{{- end -}}
{{- end -}}

{{/*
Return the Mongodb AuthDatabase
*/}}
{{- define "bcs-helm-manager.mongodb.authDatabase" -}}
{{- if .Values.mongodb.enabled }}
    {{- printf "%s" .Values.mongodb.auth.database -}}
{{- else -}}
    {{- printf "%s" .Values.externalMongo.authDatabase -}}
{{- end -}}
{{- end -}}

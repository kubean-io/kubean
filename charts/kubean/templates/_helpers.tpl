{{/*
Expand the name of the chart.
*/}}
{{- define "kubean.name" -}}
{{- default .Chart.Name .Values.kubeanOperator.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "kubean.namespace" -}}
{{- .Release.Namespace -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kubean.fullname" -}}
{{- if .Values.kubeanOperator.fullnameOverride }}
{{- .Values.kubeanOperator.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.kubeanOperator.nameOverride }}
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
{{- define "kubean.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kubean.labels" -}}
{{ include "kubean.selectorLabels" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kubean.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kubean.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app: kubean-operator
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "kubean.serviceAccountName" -}}
{{- if .Values.kubeanOperator.serviceAccount.create }}
{{- default (include "kubean.fullname" .) .Values.kubeanOperator.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.kubeanOperator.serviceAccount.name }}
{{- end }}
{{- end }}

{{- define "kubean.prehookImage" -}}
{{- printf "%s/%s:%s" .Values.sprayJob.image.registry .Values.sprayJob.image.repository (.Values.sprayJob.image.tag | default .Chart.Version) }}
{{- end }}

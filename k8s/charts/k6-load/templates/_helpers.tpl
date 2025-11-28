{{- define "k6-load.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{- define "k6-load.fullname" -}}
{{- printf "%s-%s" .Release.Name (include "k6-load.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{- define "k6-load.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
app.kubernetes.io/name: {{ include "k6-load.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "k6-load.selectorLabels" -}}
app.kubernetes.io/name: {{ include "k6-load.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

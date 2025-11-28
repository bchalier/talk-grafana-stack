{{- define "grafana-demo-app.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{- define "grafana-demo-app.fullname" -}}
{{- printf "%s-%s" .Release.Name (include "grafana-demo-app.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{- define "grafana-demo-app.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
app.kubernetes.io/name: {{ include "grafana-demo-app.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "grafana-demo-app.selectorLabels" -}}
app.kubernetes.io/name: {{ include "grafana-demo-app.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

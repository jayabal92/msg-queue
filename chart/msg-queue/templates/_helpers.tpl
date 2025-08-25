{{- define "msg-queue.name" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "msg-queue.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "msg-queue.name" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}
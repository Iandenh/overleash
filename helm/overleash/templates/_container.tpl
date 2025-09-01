{{/*
Reusable Overleash container spec
*/}}
{{- define "overleash.container" -}}
- name: {{ .Chart.Name }}
  securityContext:
    {{- toYaml .Values.securityContext | nindent 4 }}
  image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
  imagePullPolicy: {{ .Values.image.pullPolicy }}
  envFrom:
    - configMapRef:
        name: {{ include "overleash.fullname" . }}-configmap
  env:
    {{- if .Values.existingSecrets }}
    {{- toYaml .Values.existingSecrets | nindent 4 }}
    {{- end }}
    - name: OVERLEASH_LISTEN_ADDRESS
      value: ":{{ .Values.service.port }}"
  ports:
    - name: http
      containerPort: {{ .Values.service.port }}
      protocol: TCP
  livenessProbe:
    {{- toYaml .Values.livenessProbe | nindent 4 }}
  readinessProbe:
    {{- toYaml .Values.readinessProbe | nindent 4 }}
  resources:
    {{- toYaml .Values.resources | nindent 4 }}
  {{- if and (not .Values.sidecar.enabled) .Values.persistence.enabled }}
  volumeMounts:
    - mountPath: /data
      name: overleash-data
      subPath: overleash
  {{- end }}
{{- end }}
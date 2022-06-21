{{- define "common.metadata" -}}
- - group:
      - [ "apiVersion", "." ]
      - [ "name", "." ]
      - [ "namespace", "." ]
      - [ "labels" ]
      - [ "annotations" ]
    prop: metadata
{{- end }}

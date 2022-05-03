{{- define "common.metadata" -}}
- - group:
      - [ "name", "." ]
      - [ "namespace", "." ]
      - [ "labels" ]
      - [ "annotations" ]
    prop: metadata
{{- end }}
{{- define "common.metadata" -}}
- - group:
      - [ "name", "." ]
      - [ "namespace", "." ]
      - [ "labels" ]
      - [ "annotations" ]
      # resVersion 参与数据流动，但是不会展示在页面上
      - [ "resVersion" ]
    prop: metadata
{{- end }}

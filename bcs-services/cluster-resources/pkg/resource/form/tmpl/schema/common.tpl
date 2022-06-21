{{- define "common.metadata" }}
metadata:
  title: {{ i18n "基本信息" .lang }}
  type: object
  required:
    - apiVersion
    - name
    - namespace
    - labels
  properties:
    apiVersion:
      title: apiVersion
      type: string
      default: {{ .apiVersion }}
      ui:component:
        name: select
        props:
          {{- if eq .action "update" }}
          disabled: true
          {{- end }}
          placeholder: " "
          clearable: false
          remoteConfig:
            params:
              kind: {{ .kind }}
            url: "{{`{{`}} `${$context.baseUrl}/projects/${$context.projectID}/clusters/${$context.clusterID}/form_supported_api_versions` {{`}}`}}"
      ui:reactions:
        - lifetime: init
          then:
            actions:
              - "{{`{{`}} $loadDataSource {{`}}`}}"
    name:
      title: {{ i18n "名称" .lang }}
      type: string
      default: {{ .resName }}
      {{- if eq .action "update" }}
      ui:component:
        props:
          disabled: true
      {{- end }}
      ui:rules:
        - required
        - maxLength128
        - nameRegex
    namespace:
      title: {{ i18n "命名空间" .lang }}
      type: string
      default: {{ .namespace }}
      ui:component:
        name: select
        props:
          {{- if eq .action "update" }}
          disabled: true
          {{- end }}
          clearable: false
          searchable: true
          remoteConfig:
            params:
              format: selectItems
            url: "{{`{{`}} `${$context.baseUrl}/projects/${$context.projectID}/clusters/${$context.clusterID}/namespaces` {{`}}`}}"
      ui:reactions:
        - lifetime: init
          then:
            actions:
              - "{{`{{`}} $loadDataSource {{`}}`}}"
      ui:rules:
        - required
        - maxLength64
        - nameRegex
    labels:
      title: {{ i18n "标签" .lang }}
      type: array
      description: {{ i18n "将作为 Pod & Selector 标签" .lang }}
      minItems: 1
      items:
        properties:
          key:
            title: {{ i18n "键" .lang }}
            type: string
            ui:rules:
              - required
              - maxLength128
              - labelKeyRegex
          value:
            title: {{ i18n "值" .lang }}
            type: string
            ui:rules:
              - maxLength64
              - labelValRegex
        type: object
      ui:component:
        name: noTitleArray
        {{- if eq .action "update" }}
        props:
          disabled: true
        {{- end }}
    annotations:
      title: {{ i18n "注解" .lang }}
      type: array
      items:
        properties:
          key:
            title: {{ i18n "键" .lang }}
            type: string
            ui:rules:
              - required
              - maxLength128
          value:
            title: {{ i18n "值" .lang }}
            type: string
        type: object
      ui:component:
        name: noTitleArray
  ui:group:
    props:
      border: true
      showTitle: true
      type: card
{{- end }}

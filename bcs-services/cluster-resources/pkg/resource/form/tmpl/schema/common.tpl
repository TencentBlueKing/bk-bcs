{{- define "common.metadata" }}
metadata:
  title: 基本信息
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
      # 空值表示 Preferred Version
      default: ""
      ui:component:
        name: select
        props:
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
      title: 名称
      type: string
      default: {{ .resName }}
      ui:rules:
        - required
        - maxLength128
        - nameRegex
    namespace:
      title: 命名空间
      type: string
      default: {{ .namespace }}
      ui:component:
        name: select
        props:
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
      title: 标签
      type: array
      description: 将作为 Pod & Selector 标签
      minItems: 1
      items:
        properties:
          key:
            title: 键
            type: string
            ui:rules:
              - required
              - maxLength128
              - labelKeyRegex
          value:
            title: 值
            type: string
            ui:rules:
              - maxLength64
              - labelValRegex
        type: object
      ui:component:
        name: noTitleArray
    annotations:
      title: 注解
      type: array
      items:
        properties:
          key:
            title: 键
            type: string
            ui:rules:
              - required
              - maxLength128
          value:
            title: 值
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

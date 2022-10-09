{{- define "common.metadata" }}
metadata:
  title: {{ i18n "基本信息" .lang }}
  type: object
  required:
    - apiVersion
    - name
    - namespace
    # 部分资源类型允许不填写 labels
    {{- if isLabelRequired .kind }}
    - labels
    {{- end }}
  properties:
    apiVersion:
      title: apiVersion
      type: string
      default: {{ .apiVersion }}
      ui:component:
        name: select
        props:
          # 更新时候不允许编辑 APIVersion
          disabled: {{ eq .action "update" }}
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
      # 更新时候不允许编辑资源名称
      ui:component:
        props:
          disabled: {{ eq .action "update" }}
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
          visible: {{ isNSRequired .kind }}
          # 更新时候不允许编辑命名空间
          disabled: {{ eq .action "update" }}
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
      minItems: {{ if isLabelRequired .kind }} 1 {{ else }} 0 {{ end }}
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
        name: bfArray
        props:
          visible: {{ isLabelVisible .kind }}
          disabled: {{ eq .action "update" }}
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
            ui:reactions:
              - if: "{{`{{`}} $self.value === 'io.tencent.bcs.editFormat' {{`}}`}}"
                then:
                  state:
                    disabled: true
                else:
                  state:
                    disabled: false
              - target: "{{`{{`}} $widgetNode?.getSibling('value')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'io.tencent.bcs.editFormat' {{`}}`}}"
                then:
                  state:
                    disabled: true
                else:
                  state:
                    disabled: false
          value:
            title: {{ i18n "值" .lang }}
            type: string
        type: object
      ui:component:
        name: bfArray
        props:
          visible: {{ isAnnoVisible .kind }}
  ui:group:
    props:
      border: true
      showTitle: true
      type: card
      hideEmptyRow: true
{{- end }}

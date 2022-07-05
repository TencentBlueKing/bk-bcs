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
          # 更新时候不允许编辑 APIVersion
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
      # 更新时候不允许编辑资源名称
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
          # 更新时候不允许编辑命名空间
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
      {{- if eq .kind "HookTemplate" }}
      minItems: 0
      {{- else }}
      minItems: 1
      {{- end }}
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
        # TODO 如果后续 common.tpl 对资源类型的定制增多的话，可以考虑封装成方法
        # HookTemplate 类型资源不展示 labels
        {{- if eq .kind "HookTemplate" }}
        props:
          visible: false
        {{- else if eq .action "update" }}
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
        name: noTitleArray
        {{- if eq .kind "HookTemplate" }}
        props:
          visible: false
        {{- end }}
  ui:group:
    props:
      border: true
      showTitle: true
      type: card
{{- end }}

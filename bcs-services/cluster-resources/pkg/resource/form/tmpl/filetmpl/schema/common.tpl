{{- define "common.metadata" }}
metadata:
  title: {{ i18n "基本信息" .lang }}
  type: object
  properties:
    name:
      title: {{ i18n "名称" .lang }}
      type: string
      default: {{ .resName }}
      # 更新时候不允许编辑资源名称
      ui:component:
        props:
          disabled: {{ eq .action "update" }}
          visible: false
    labels:
      title: {{ i18n "标签" .lang }}
      type: array
      default: {{ .selectorLabel }}
      items:
        properties:
          key:
            title: {{ i18n "键" .lang }}
            type: string
            ui:rules:
              - required
              - maxLength128
              - labelKeyRegexWithVar
            ui:reactions:
              - if: "{{`{{`}} $self.value === 'workload.bcs.tencent.io/workloadSelector' {{`}}`}}"
                then:
                  state:
                    disabled: true
                else:
                  state:
                    disabled: false
              - target: "{{`{{`}} $widgetNode?.getSibling('value')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'workload.bcs.tencent.io/workloadSelector' {{`}}`}}"
                then:
                  state:
                    disabled: true
                else:
                  state:
                    disabled: false
          value:
            title: {{ i18n "值" .lang }}
            type: string
            ui:rules:
              - maxLength64
              - labelValRegexWithVar
        type: object
      ui:component:
        name: bfArray
        props:
          visible: {{ isLabelVisible .kind }}
          disabled: {{ isLabelEditDisabled .kind .action }}
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
              - if: "{{`{{`}} $self.value === 'io.tencent.bcs.editFormat' || $self.value === 'io.tencent.bcs.labelSelected' || $self.value === 'io.tencent.paas.creator' || $self.value === 'io.tencent.paas.updator' {{`}}`}}"
                then:
                  state:
                    disabled: true
                else:
                  state:
                    disabled: false
              - target: "{{`{{`}} $widgetNode?.getSibling('value')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'io.tencent.bcs.editFormat' || $self.value === 'io.tencent.bcs.labelSelected' || $self.value === 'io.tencent.paas.creator' || $self.value === 'io.tencent.paas.updator' {{`}}`}}"
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
    resVersion:
      type: string
      # resVersion 在更新时会参与数据流动，但是不允许用户编辑
      ui:component:
        props:
          visible: false
          disabled: true
  ui:group:
    props:
      border: false
      showTitle: false
      type: card
      hideEmptyRow: true
{{- end }}

{{- define "affinity.podAffinity" }}
podAffinity:
  title: {{ i18n "Pod 规则" .lang }}
  type: array
  items:
    type: object
    required:
      - topologyKey
    properties:
      namespaces:
        title: {{ i18n "命名空间" .lang }}
        type: array
        ui:component:
          name: select
          props:
            clearable: false
      priority:
        title: {{ i18n "优先级" .lang }}
        type: string
        default: preferred
        ui:component:
          name: radio
          props:
            datasource:
              - label: {{ i18n "优先" .lang }}
                value: preferred
              - label: {{ i18n "必须" .lang }}
                value: required
        ui:reactions:
          - target: "{{`{{`}} $widgetNode?.getSibling('weight')?.id {{`}}`}}"
            if: "{{`{{`}} $self.value === 'required' {{`}}`}}"
            then:
              state:
                visible: false
            else:
              state:
                visible: true
      selector:
        type: object
        properties:
          expressions:
            type: array
            items:
              properties:
                key:
                  title: key
                  type: string
                  ui:rules:
                    - required
                    - maxLength128
                op:
                  title: op
                  type: string
                  ui:component:
                    name: select
                    props:
                      datasource:
                        - label: Exists
                          value: Exists
                        - label: DoesNotExist
                          value: DoesNotExist
                        - label: In
                          value: In
                        - label: NotIn
                          value: NotIn
                  ui:reactions:
                    - target: "{{`{{`}} $widgetNode?.getSibling('values')?.id {{`}}`}}"
                      if: "{{`{{`}} $self.value === 'Exists' || $self.value === 'DoesNotExist' {{`}}`}}"
                      then:
                        state:
                          disabled: true
                          value: ""
                      else:
                        state:
                          disabled: false
                values:
                  title: values
                  type: string
                  ui:component:
                    props:
                      placeholder: {{ i18n "值（多个值请以英文逗号分隔）" .lang }}
                  ui:rules:
                    - maxLength128
              type: object
            ui:component:
              name: bfArray
          labels:
            type: array
            items:
              properties:
                key:
                  title: key
                  type: string
                  ui:rules:
                    - required
                    - maxLength128
                value:
                  title: value
                  type: string
                  ui:rules:
                    - maxLength128
              type: object
            ui:component:
              name: bfArray
      topologyKey:
        title: {{ i18n "拓扑键" .lang }}
        type: string
        ui:rules:
          - required
          - maxLength250
      type:
        title: {{ i18n "类型" .lang }}
        type: string
        default: affinity
        ui:component:
          name: radio
          props:
            datasource:
              - label: {{ i18n "亲和性" .lang }}
                value: affinity
              - label: {{ i18n "反亲和性" .lang }}
                value: antiAffinity
      weight:
        title: {{ i18n "权重" .lang }}
        type: integer
        default: 10
        ui:component:
          props:
            max: 100
            min: 1
    ui:group:
      props:
        showTitle: false
        type: normal
      style:
        background: '#fff'
  ui:group:
    props:
      showTitle: true
      type: card
    style:
      background: '#F5F7FA'
{{- end }}

{{- define "affinity.nodeAffinity" }}
nodeAffinity:
  title: {{ i18n "Node 规则" .lang }}
  type: array
  items:
    type: object
    properties:
      priority:
        title: {{ i18n "优先级" .lang }}
        type: string
        default: preferred
        ui:component:
          name: radio
          props:
            datasource:
              - label: {{ i18n "优先" .lang }}
                value: preferred
              - label: {{ i18n "必须" .lang }}
                value: required
        ui:reactions:
          - target: "{{`{{`}} $widgetNode?.getSibling('weight')?.id {{`}}`}}"
            if: "{{`{{`}} $self.value === 'required' {{`}}`}}"
            then:
              state:
                visible: false
            else:
              state:
                visible: true
      selector:
        type: object
        properties:
          expressions:
            items:
              properties:
                key:
                  title: key
                  type: string
                  ui:rules:
                    - required
                    - maxLength128
                op:
                  title: op
                  type: string
                  ui:component:
                    name: select
                    props:
                      datasource:
                        - label: Lt
                          value: Lt
                        - label: Gt
                          value: Gt
                        - label: Exists
                          value: Exists
                        - label: DoesNotExist
                          value: DoesNotExist
                        - label: In
                          value: In
                        - label: NotIn
                          value: NotIn
                  ui:reactions:
                    - target: "{{`{{`}} $widgetNode?.getSibling('values')?.id {{`}}`}}"
                      if: "{{`{{`}} $self.value === 'Exists' || $self.value === 'DoesNotExist' {{`}}`}}"
                      then:
                        state:
                          disabled: true
                          value: ""
                      else:
                        state:
                          disabled: false
                values:
                  title: values
                  type: string
                  ui:component:
                    props:
                      placeholder: {{ i18n "值（多个值请以英文逗号分隔）" .lang }}
                  ui:rules:
                    - maxLength128
              type: object
            type: array
            ui:component:
              name: bfArray
          fields:
            items:
              properties:
                key:
                  title: key
                  type: string
                  ui:rules:
                    - required
                    - maxLength128
                op:
                  title: op
                  type: string
                  ui:component:
                    name: select
                    props:
                      datasource:
                        - label: Lt
                          value: Lt
                        - label: Gt
                          value: Gt
                        - label: Exists
                          value: Exists
                        - label: DoesNotExist
                          value: DoesNotExist
                        - label: In
                          value: In
                        - label: NotIn
                          value: NotIn
                  ui:reactions:
                    - target: "{{`{{`}} $widgetNode?.getSibling('values')?.id {{`}}`}}"
                      if: "{{`{{`}} $self.value === 'Exists' || $self.value === 'DoesNotExist' {{`}}`}}"
                      then:
                        state:
                          disabled: true
                          value: ""
                      else:
                        state:
                          disabled: false
                values:
                  title: values
                  type: string
                  ui:component:
                    props:
                      placeholder: {{ i18n "值（多个值请以英文逗号分隔）" .lang }}
                  ui:rules:
                    - maxLength128
              type: object
            type: array
            ui:component:
              name: bfArray
      weight:
        default: 10
        title: {{ i18n "权重" .lang }}
        type: integer
        ui:component:
          props:
            max: 100
            min: 1
    ui:group:
      props:
        showTitle: false
        type: normal
      style:
        background: '#fff'
  ui:group:
    props:
      showTitle: true
      type: card
    style:
      background: '#F5F7FA'
{{- end }}

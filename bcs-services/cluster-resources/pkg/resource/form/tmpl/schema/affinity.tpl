{{- define "affinity.podAffinity" }}
podAffinity:
  title: Pod 规则
  type: array
  items:
    type: object
    properties:
      namespaces:
        title: 命名空间
        type: array
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
      priority:
        title: 优先级
        type: string
        default: preferred
        ui:component:
          name: radio
          props:
            datasource:
              - label: 优先
                value: preferred
              - label: 必须
                value: required
        ui:reactions:
          - target: "{{`{{`}} $widgetNode?.getSibling('weight')?.id {{`}}`}}"
            if: "{{`{{`}} $self.value === 'required' {{`}}`}}"
            then:
              state:
                disabled: true
            else:
              state:
                disabled: false
      selector:
        type: object
        properties:
          expressions:
            title: matchExpressions
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
                values:
                  title: values
                  type: string
                  ui:rules:
                    - maxLength128
              type: object
            ui:component:
              name: noTitleArray
          labels:
            title: matchFields
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
              name: noTitleArray
      topologyKey:
        title: 拓扑键
        type: string
        ui:rules:
          - maxLength250
      type:
        title: 类型
        type: string
        default: affinity
        ui:component:
          name: radio
          props:
            datasource:
              - label: 亲和性
                value: affinity
              - label: 反亲和性
                value: antiAffinity
      weight:
        title: 权重
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
  title: Node 规则
  type: array
  items:
    type: object
    properties:
      priority:
        title: 优先级
        type: string
        default: preferred
        ui:component:
          name: radio
          props:
            datasource:
              - label: 优先
                value: preferred
              - label: 必须
                value: required
        ui:reactions:
          - target: "{{`{{`}} $widgetNode?.getSibling('weight')?.id {{`}}`}}"
            if: "{{`{{`}} $self.value === 'required' {{`}}`}}"
            then:
              state:
                disabled: true
            else:
              state:
                disabled: false
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
                values:
                  title: values
                  type: string
                  ui:rules:
                    - maxLength128
              type: object
            title: matchExpressions
            type: array
            ui:component:
              name: noTitleArray
          labels:
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
                values:
                  title: values
                  type: string
                  ui:rules:
                    - maxLength128
              type: object
            title: matchFields
            type: array
            ui:component:
              name: noTitleArray
      weight:
        default: 10
        title: 权重
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

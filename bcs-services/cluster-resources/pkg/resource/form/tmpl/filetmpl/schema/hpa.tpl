{{- define "hpa.refObj" }}
ref:
  title: {{ i18n "关联对象" .lang }}
  type: object
  required:
    - kind
    - apiVersion
    - resName
    - minReplicas
    - maxReplicas
  properties:
    kind:
      title: {{ i18n "资源类型" .lang }}
      type: string
      default: Deployment
      ui:component:
        name: select
        props:
          clearable: false
          datasource:
            - label: Deployment
              value: Deployment
            - label: StatefulSet
              value: StatefulSet
            - label: GameDeployment
              value: GameDeployment
            - label: GameStatefulSet
              value: GameStatefulSet
      ui:reactions:
        - target: spec.ref.apiVersion
          if: "{{`{{`}} $self.value === 'Deployment' || self.value == 'StatefulSet' {{`}}`}}"
          then:
            state:
              value: "apps/v1"
        - target: spec.ref.apiVersion
          if: "{{`{{`}} $self.value === 'GameDeployment' || self.value == 'GameStatefulSet' {{`}}`}}"
          then:
            state:
              value: "tkex.tencent.com/v1alpha1"
    apiVersion:
      title: apiVersion
      type: string
      ui:component:
        props:
          # 目前 HPA 关联资源的 APIVersion 不需要用户关心，但是需要参与数据流动，因此做界面上的隐藏
          visible: false
      ui:rules:
        - required
        - maxLength128
    resName:
      title: {{ i18n "资源名称" .lang }}
      type: string
      default: deployment-test
      ui:rules:
        - required
    minReplicas:
      title: {{ i18n "最小副本数" .lang }}
      type: integer
      default: 1
      ui:component:
        props:
          max: 4096
    maxReplicas:
      title: {{ i18n "最大副本数" .lang }}
      type: integer
      default: 10
      ui:component:
        props:
          max: 4096
      ui:rules:
        - validator: "{{`{{`}} $self.getValue('spec.ref.minReplicas') <= $self.value {{`}}`}}"
          message: {{ i18n "最大副本数必须大于最小副本数" .lang }}
{{- end }}

{{- define "hpa.resMetric" }}
resource:
  title: {{ i18n "Resource 指标" .lang }}
  type: object
  properties:
    items:
      type: array
      items:
        type: object
        required:
          - name
          - type
          - percent
          - cpuVal
          - memVal
        properties:
          name:
            title: {{ i18n "资源" .lang }}
            type: string
            default: cpu
            ui:component:
              name: select
              props:
                clearable: false
                datasource:
                  - label: CPU
                    value: cpu
                  - label: Memory
                    value: memory
            ui:reactions:
              # 仅当资源类型为 cpu 且指标类型为 AverageValue 时，展示 mCPUs 为单位的输入框
              - target: "{{`{{`}} $widgetNode?.getSibling('cpuVal')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'cpu' && $widgetNode?.getSibling('type')?.instance?.value === 'AverageValue' {{`}}`}}"
                then:
                  state:
                    visible: true
                else:
                  state:
                    visible: false
              # 仅当资源类型为 memory 且指标类型为 AverageValue 时，展示 Mi 为单位的输入框
              - target: "{{`{{`}} $widgetNode?.getSibling('memVal')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'memory' && $widgetNode?.getSibling('type')?.instance?.value === 'AverageValue' {{`}}`}}"
                then:
                  state:
                    visible: true
                else:
                  state:
                    visible: false
          type:
            title: {{ i18n "指标类型" .lang }}
            type: string
            default: Utilization
            ui:component:
              name: select
              props:
                clearable: false
                datasource:
                  - label: AverageValue
                    value: AverageValue
                    disabled: false
                    tips: {{ i18n "工作负载资源使用绝对数值" .lang }}
                  - label: AverageUtilization
                    value: Utilization
                    disabled: false
                    tips: {{ i18n "工作负载所有 Pod 资源实际使用值 / 资源 Request 值" .lang }}
            ui:reactions:
              # 仅当指标类型为 Utilization 时，展示 % 为单位的输入框
              - target: "{{`{{`}} $widgetNode?.getSibling('percent')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'Utilization' {{`}}`}}"
                then:
                  state:
                    visible: true
                else:
                  state:
                    visible: false
              # 仅当指标类型为 AverageValue 且资源类型为 cpu 时，展示 mCPUs 为单位的输入框
              - target: "{{`{{`}} $widgetNode?.getSibling('cpuVal')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'AverageValue' && $widgetNode?.getSibling('name')?.instance?.value === 'cpu' {{`}}`}}"
                then:
                  state:
                    visible: true
                else:
                  state:
                    visible: false
              # 仅当指标类型为 AverageValue 且资源类型为 memory 时，展示 Mi 为单位的输入框
              - target: "{{`{{`}} $widgetNode?.getSibling('memVal')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'AverageValue' && $widgetNode?.getSibling('name')?.instance?.value === 'memory' {{`}}`}}"
                then:
                  state:
                    visible: true
                else:
                  state:
                    visible: false
          percent:
            title: {{ i18n "值" .lang }}
            type: integer
            default: 80
            ui:component:
              name: bfInput
              props:
                max: 100
                unit: "%"
            ui:rules:
              - validator: "{{`{{`}} $widgetNode?.getSibling('type')?.instance?.value !== 'Utilization' || $self.value {{`}}`}}"
                message: {{ i18n "值不能为零或空" .lang }}
          cpuVal:
            title: {{ i18n "值" .lang }}
            type: integer
            default: 2000
            ui:component:
              name: bfInput
              props:
                max: 256000
                unit: mCPUs
            ui:rules:
              - validator: "{{`{{`}} $widgetNode?.getSibling('type')?.instance?.value !== 'AverageValue' || $widgetNode?.getSibling('name')?.instance?.value !== 'cpu' || $self.value {{`}}`}}"
                message: {{ i18n "值不能为零或空" .lang }}
          memVal:
            title: {{ i18n "值" .lang }}
            type: integer
            default: 1024
            ui:component:
              name: bfInput
              props:
                max: 256000
                unit: MiB
            ui:rules:
              - validator: "{{`{{`}} $widgetNode?.getSibling('type')?.instance?.value !== 'AverageValue' || $widgetNode?.getSibling('name')?.instance?.value !== 'memory' || $self.value {{`}}`}}"
                message: {{ i18n "值不能为零或空" .lang }}
        ui:group:
          props:
            showTitle: false
            type: normal
          style:
            background: '#F5F7FA'
{{- end }}

# HPA 的 ContainerResource 先不启用，两点原因：
#  1. 需要 kubelet 1.20 支持，且目前还是 alpha 版本
#  2. 目前 KUBE_FEATURE_GATES 的确没有开启这个特性
{{- define "hpa.containerResMetric" }}
containerRes:
  title: {{ i18n "ContainerResource 指标" .lang }}
  type: object
  properties:
    items:
      type: array
      items:
        type: object
        required:
          - name
          - containerName
          - type
          - value
        properties:
          name:
            title: {{ i18n "指标名称" .lang }}
            type: string
            default: cpu
            ui:component:
              name: select
              props:
                clearable: false
                datasource:
                  - label: CPU
                    value: cpu
                  - label: Memory
                    value: memory
          containerName:
            title: {{ i18n "容器名称" .lang }}
            type: string
            ui:rules:
              - required
              - maxLength128
              - nameRegexWithVar
          type:
            title: {{ i18n "指标类型" .lang }}
            type: string
            default: AverageValue
            ui:component:
              name: select
              props:
                clearable: false
                datasource:
                  - label: AverageValue
                    value: AverageValue
                  - label: AverageUtilization
                    value: Utilization
          value:
            title: {{ i18n "值" .lang }}
            type: string
            default: "80"
            ui:rules:
              - required
              - maxLength64
        ui:group:
          props:
            showTitle: false
            type: normal
          style:
            background: '#F5F7FA'
{{- end }}

{{- define "hpa.externalMetric" }}
external:
  title: {{ i18n "External 指标" .lang }}
  type: object
  properties:
    items:
      type: array
      items:
        type: object
        required:
          - name
          - type
          - value
        properties:
          name:
            title: {{ i18n "指标名称" .lang }}
            type: string
            ui:rules:
              - required
              - maxLength128
          type:
            title: {{ i18n "指标类型" .lang }}
            type: string
            default: AverageValue
            ui:component:
              name: select
              props:
                clearable: false
                datasource:
                  - label: AverageValue
                    value: AverageValue
                  - label: Value
                    value: Value
          value:
            title: {{ i18n "值" .lang }}
            type: string
            default: "10"
            ui:rules:
              - required
              - maxLength64
          selector:
            title: {{ i18n "选择器" .lang }}
            type: object
            properties:
              expressions:
                type: array
                items:
                  type: object
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
                            - label: In
                              value: In
                            - label: NotIn
                              value: NotIn
                            - label: Exists
                              value: Exists
                            - label: DoesNotExist
                              value: DoesNotExist
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
        ui:group:
          props:
            showTitle: false
            type: normal
          style:
            background: '#F5F7FA'
{{- end }}

{{- define "hpa.objMetric" }}
object:
  title: {{ i18n "Object 指标" .lang }}
  type: object
  properties:
    items:
      type: array
      items:
        type: object
        required:
          - name
          - kind
          - apiVersion
          - resName
          - type
          - value
        properties:
          name:
            title: {{ i18n "指标名称" .lang }}
            type: string
            ui:rules:
              - required
              - maxLength128
          kind:
            title: {{ i18n "资源类型" .lang }}
            type: string
            default: Deployment
            ui:component:
              name: select
              props:
                clearable: false
                datasource:
                  - label: Deployment
                    value: Deployment
                  - label: StatefulSet
                    value: StatefulSet
                  - label: GameDeployment
                    value: GameDeployment
                  - label: GameStatefulSet
                    value: GameStatefulSet
            ui:reactions:
              - target: "{{`{{`}} $widgetNode?.getSibling('apiVersion')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'Deployment' || self.value == 'StatefulSet' {{`}}`}}"
                then:
                  state:
                    value: "apps/v1"
              - target: "{{`{{`}} $widgetNode?.getSibling('apiVersion')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'GameDeployment' || self.value == 'GameStatefulSet' {{`}}`}}"
                then:
                  state:
                    value: "tkex.tencent.com/v1alpha1"
          apiVersion:
            title: apiVersion
            type: string
            ui:component:
              props:
                # 目前 HPA 关联资源的 APIVersion 不需要用户关心，但是需要参与数据流动，因此做界面上的隐藏
                visible: false
            ui:rules:
              - required
              - maxLength128
          resName:
            title: {{ i18n "资源名称" .lang }}
            type: string
            ui:rules:
              - required
              - maxLength128
              - nameRegexWithVar
          type:
            title: {{ i18n "指标类型" .lang }}
            type: string
            default: AverageValue
            ui:component:
              name: select
              props:
                clearable: false
                datasource:
                  - label: AverageValue
                    value: AverageValue
                  - label: Value
                    value: Value
          value:
            title: {{ i18n "值" .lang }}
            type: string
            default: "10"
            ui:rules:
              - required
              - maxLength64
          selector:
            title: {{ i18n "选择器" .lang }}
            type: object
            properties:
              expressions:
                type: array
                items:
                  type: object
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
                            - label: In
                              value: In
                            - label: NotIn
                              value: NotIn
                            - label: Exists
                              value: Exists
                            - label: DoesNotExist
                              value: DoesNotExist
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
        ui:group:
          props:
            showTitle: false
            type: normal
          style:
            background: '#F5F7FA'
{{- end }}

{{- define "hpa.podMetric" }}
pod:
  title: {{ i18n "Pod 指标" .lang }}
  type: object
  properties:
    items:
      type: array
      items:
        type: object
        required:
          - name
          - type
          - value
        properties:
          name:
            title: {{ i18n "指标名称" .lang }}
            type: string
            ui:rules:
              - required
              - maxLength128
          type:
            title: {{ i18n "指标类型" .lang }}
            type: string
            default: AverageValue
            ui:component:
              name: select
              props:
                clearable: false
                datasource:
                  - label: AverageValue
                    value: AverageValue
          value:
            title: {{ i18n "值" .lang }}
            type: string
            default: "10"
            ui:rules:
              - required
              - maxLength64
          selector:
            title: {{ i18n "选择器" .lang }}
            type: object
            properties:
              expressions:
                type: array
                items:
                  type: object
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
                            - label: In
                              value: In
                            - label: NotIn
                              value: NotIn
                            - label: Exists
                              value: Exists
                            - label: DoesNotExist
                              value: DoesNotExist
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
        ui:group:
          props:
            showTitle: false
            type: normal
          style:
            background: '#F5F7FA'
{{- end }}

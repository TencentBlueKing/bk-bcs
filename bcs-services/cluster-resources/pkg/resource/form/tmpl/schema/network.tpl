{{- define "network.ingRules" }}
ruleConf:
  title: {{ i18n "规则" .lang }}
  type: object
  properties:
    rules:
      type: array
      items:
        type: object
        properties:
          domain:
            title: {{ i18n "域名" .lang }}
            type: string
            ui:rules:
              - maxLength128
          paths:
            title: {{ i18n "路径" .lang }}
            type: array
            minItems: 1
            items:
              type: object
              properties:
                type:
                  title: {{ i18n "类型" .lang }}
                  type: string
                  default: ImplementationSpecific
                  ui:component:
                    name: select
                    props:
                      clearable: false
                      datasource:
                        - label: Prefix
                          value: Prefix
                        - label: Exact
                          value: Exact
                        - label: ImplementationSpecific
                          value: ImplementationSpecific
                  # 仅当 apiVersion 为 networking.k8s.io/v1 才能使用 pathType
                  ui:reactions:
                    - lifetime: init
                      if: "{{`{{`}} $self.getValue('metadata.apiVersion') === 'networking.k8s.io/v1' {{`}}`}}"
                      then:
                        state:
                          disabled: false
                      else:
                        state:
                          disabled: true
                          value: ""
                    - source: "metadata.apiVersion"
                      if: "{{`{{`}} $self.getValue('metadata.apiVersion') === 'networking.k8s.io/v1' {{`}}`}}"
                      then:
                        state:
                          disabled: false
                      else:
                        state:
                          disabled: true
                          value: ""
                path:
                  title: {{ i18n "路径" .lang }}
                  type: string
                  ui:rules:
                    - required
                    - maxLength128
                    - validator: "{{`{{`}} $self.value.startsWith('/') {{`}}`}}"
                      message: {{ i18n "需要为绝对路径" .lang }}
                targetSVC:
                  title: {{ i18n "目标 Service" .lang }}
                  type: string
                  ui:component:
                    name: select
                    props:
                      clearable: true
                      searchable: true
                      remoteConfig:
                        params:
                          format: selectItems
                        url: "{{`{{`}} `${$context.baseUrl}/projects/${$context.projectID}/clusters/${$context.clusterID}/namespaces/${$self.getValue('metadata.namespace')}/networks/services` {{`}}`}}"
                  ui:reactions:
                    - lifetime: init
                      then:
                        actions:
                          - "{{`{{`}} $loadDataSource {{`}}`}}"
                    - source: "metadata.namespace"
                      then:
                        state:
                          value: ""
                        actions:
                          - "{{`{{`}} $loadDataSource {{`}}`}}"
                  ui:rules:
                    - required
                port:
                  title: {{ i18n "端口" .lang }}
                  type: integer
                  ui:component:
                    props:
                      min: 1
                      max: 65535
                  ui:rules:
                    - validator: "{{`{{`}} $self.value {{`}}`}}"
                      message: {{ i18n "值不能为空" .lang }}
            ui:component:
              name: bfArray
            ui:props:
              showTitle: false
        ui:group:
          props:
            type: normal
          style:
            background: '#F5F7FA'
{{- end }}

{{- define "network.ingNetwork" }}
network:
  title: {{ i18n "网络" .lang }}
  type: object
  required:
    - existLBID
    - subNetID
  properties:
    clbUseType:
      title: {{ i18n "CLB 使用方式" .lang }}
      type: string
      default: "useExists"
      description: {{ i18n "已存在的 CLB 实例：需先手工创建 CLB 实例，Ingress 实例删除时 CLB 实例不会被删除，重建 Ingress 实例后 VIP 不会发生变化<br>自动新建 CLB 实例：无需先手工创建 CLB 实例，Ingress 实例删除时 CLB 实例会随之删除，重建 Ingress 实例后 VIP 会发生变化" .lang | quote }}
      ui:component:
        name: select
        props:
          clearable: false
          datasource:
            - label: {{ i18n "已存在的 CLB 实例" .lang }}
              value: useExists
            - label: {{ i18n "自动新建 CLB 实例" .lang }}
              value: autoCreate
      ui:reactions:
        - target: spec.network.existLBID
          if: "{{`{{`}} $self.value === 'useExists' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
        - target: spec.network.clbType
          if: "{{`{{`}} $self.value === 'autoCreate' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
        - target: spec.network.subNetID
          if: "{{`{{`}} $self.value === 'autoCreate' && $self.getValue('spec.network.clbType') === 'internal' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
    existLBID:
      title: "CLB ID"
      type: string
      ui:rules:
        # 不是使用"已存在 CLB 实例"时 或 非 qcloud 类型 ingress，可以不填该值
        - validator: "{{`{{`}} $self.getValue('controller.type') !== 'qcloud' || $self.getValue('spec.network.clbUseType') !== 'useExists' || $self.value {{`}}`}}"
          message: {{ i18n "值不能为空" .lang }}
    clbType:
      title: {{ i18n "CLB 类型" .lang }}
      type: string
      default: external
      ui:component:
        name: radio
        props:
          datasource:
            - label: {{ i18n "外网 CLB" .lang }}
              value: external
            - label: {{ i18n "内网 CLB" .lang }}
              value: internal
      ui:reactions:
        - target: spec.network.subNetID
          if: "{{`{{`}} $self.value === 'internal' && $self.getValue('spec.network.clbUseType') === 'autoCreate' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
    subNetID:
      title: {{ i18n "子网 ID" .lang }}
      type: string
      ui:rules:
        # ingress 为 qcloud 类型 且 "自动创建 CLB 实例" 且 "为内网 CLB" 时，必须填值
        - validator: "{{`{{`}} $self.getValue('controller.type') !== 'qcloud' || $self.getValue('spec.network.clbUseType') !== 'autoCreate' || $self.getValue('spec.network.clbType') !== 'internal' || $self.value {{`}}`}}"
          message: {{ i18n "值不能为空" .lang }}
{{- end }}

{{- define "network.ingDefaultBackend" }}
defaultBackend:
  title: {{ i18n "默认后端" .lang }}
  type: object
  properties:
    targetSVC:
      title: {{ i18n "目标 Service" .lang }}
      type: string
      default: ""
      ui:component:
        name: select
        props:
          clearable: true
          searchable: true
          remoteConfig:
            params:
              format: selectItems
            url: "{{`{{`}} `${$context.baseUrl}/projects/${$context.projectID}/clusters/${$context.clusterID}/namespaces/${$self.getValue('metadata.namespace')}/networks/services` {{`}}`}}"
      ui:reactions:
        - lifetime: init
          then:
            actions:
              - "{{`{{`}} $loadDataSource {{`}}`}}"
        - source: "metadata.namespace"
          then:
            state:
              value: ""
            actions:
              - "{{`{{`}} $loadDataSource {{`}}`}}"
      ui:rules:
        - maxLength128
    port:
      title: {{ i18n "端口" .lang }}
      type: integer
      ui:component:
        props:
          min: 1
          max: 65535
      ui:rules:
        # 没有选择默认后后端 svc 时，可以不填端口值
        - validator: "{{`{{`}} $self.getValue('spec.defaultBackend.targetSVC') == '' || $self.value {{`}}`}}"
          message: {{ i18n "值不能为空" .lang }}
{{- end }}

{{- define "network.ingCert" }}
cert:
  title: {{ i18n "证书" .lang }}
  type: object
  properties:
    tls:
      type: array
      items:
        type: object
        properties:
          secretName:
            title: {{ i18n "证书" .lang }}
            type: string
            ui:component:
              name: select
              props:
                clearable: true
                searchable: true
                remoteConfig:
                  params:
                    format: selectItems
                  url: "{{`{{`}} `${$context.baseUrl}/projects/${$context.projectID}/clusters/${$context.clusterID}/namespaces/${$self.getValue('metadata.namespace')}/configs/secrets` {{`}}`}}"
            ui:reactions:
              - lifetime: init
                then:
                  actions:
                    - "{{`{{`}} $loadDataSource {{`}}`}}"
              - source: "metadata.namespace"
                then:
                  actions:
                    - "{{`{{`}} $loadDataSource {{`}}`}}"
          hosts:
            title: Hosts
            type: array
            items:
              type: string
              ui:rules:
                - required
                - maxLength128
            ui:component:
              name: bfArray
        ui:group:
          props:
            type: normal
          style:
            background: '#F5F7FA'
{{- end }}

{{- define "network.svcPort" }}
portConf:
  title: {{ i18n "端口配置" .lang }}
  type: object
  properties:
    type:
      title: {{ i18n "类型" .lang }}
      type: string
      default: ClusterIP
      ui:component:
        name: select
        props:
          clearable: false
          datasource:
            - label: ClusterIP
              value: ClusterIP
            - label: NodePort
              value: NodePort
            - label: LoadBalancer
              value: LoadBalancer
    ports:
      type: array
      minItems: 1
      items:
        type: object
        properties:
          name:
            title: {{ i18n "名称" .lang }}
            type: string
            ui:rules:
              - required
              - maxLength64
          port:
            title: {{ i18n "监听端口" .lang }}
            type: integer
            ui:component:
              props:
                min: 1
                max: 65535
            ui:rules:
              - validator: "{{`{{`}} $self.value {{`}}`}}"
                message: {{ i18n "值不能为空" .lang }}
          protocol:
            title: {{ i18n "协议" .lang }}
            type: string
            default: TCP
            ui:component:
              name: select
              props:
                clearable: false
                datasource:
                  - label: TCP
                    value: TCP
                  - label: UDP
                    value: UDP
          targetPort:
            title: {{ i18n "目标端口" .lang }}
            type: integer
            ui:component:
              props:
                min: 1
                max: 65535
            ui:rules:
              - validator: "{{`{{`}} $self.value {{`}}`}}"
                message: {{ i18n "值不能为空" .lang }}
          nodePort:
            title: {{ i18n "节点端口" .lang }}
            type: integer
            ui:component:
              props:
                min: 30000
                max: 32767
            ui:reactions:
              - lifetime: init
                if: "{{`{{`}} $self.getValue('spec.portConf.type') === 'ClusterIP' {{`}}`}}"
                then:
                  state:
                    disabled: true
                else:
                  state:
                    disabled: false
              - source: "spec.portConf.type"
                if: "{{`{{`}} $self.getValue('spec.portConf.type') === 'ClusterIP' {{`}}`}}"
                then:
                  state:
                    disabled: true
                else:
                  state:
                    disabled: false
      ui:component:
        name: bfArray
      ui:props:
        showTitle: false
{{- end }}

{{- define "network.svcSelector" }}
selector:
  title: {{ i18n "选择器" .lang }}
  type: object
  properties:
    labels:
      type: array
      items:
        type: object
        properties:
          key:
            title: {{ i18n "键" .lang }}
            type: string
            ui:rules:
              - required
              - labelKeyRegex
              - maxLength128
          value:
            title: {{ i18n "值" .lang }}
            type: string
            ui:rules:
              - maxLength128
      ui:component:
        name: bfArray
      ui:props:
        showTitle: false
{{- end }}

{{- define "network.sessionAffinity" }}
sessionAffinity:
  title: {{ i18n "Session 亲和性" .lang }}
  type: object
  properties:
    type:
      title: {{ i18n "类型" .lang }}
      type: string
      default: None
      ui:component:
        name: radio
        props:
          datasource:
            - label: None
              value: None
            - label: ClientIP
              value: ClientIP
      ui:reactions:
        - target: "{{`{{`}} $widgetNode?.getSibling('stickyTime')?.id {{`}}`}}"
          if: "{{`{{`}} $self.value === 'ClientIP' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
    stickyTime:
      title: "SessionStickyTime"
      type: integer
      default: 120
      ui:component:
        name: bfInput
        props:
          max: 86400
          unit: s
{{- end }}

{{- define "network.svcIPConf" }}
ip:
  title: {{ i18n "IP 地址" .lang }}
  type: object
  properties:
    address:
      title: {{ i18n "IP 地址" .lang }}
      type: string
      ui:rules:
        - maxLength128
    external:
      title: {{ i18n "外部 IP" .lang }}
      type: array
      items:
        type: string
        ui:rules:
          - required
          - maxLength128
      ui:component:
        name: bfArray
{{- end }}

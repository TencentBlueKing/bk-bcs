{{- define "network.ingRules" }}
ruleConf:
  title: {{ i18n "规则" .lang }}
  type: object
  properties:
    rules:
      type: array
      minItems: 1
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
                  default: /testpath
                  ui:rules:
                    - required
                    - maxLength128
                    - validator: "{{`{{`}} $self.value.startsWith('/') {{`}}`}}"
                      message: {{ i18n "需要为绝对路径" .lang }}
                targetSVC:
                  title: {{ i18n "目标 Service" .lang }}
                  type: string
                  default: test
                  ui:rules:
                    - required
                port:
                  title: {{ i18n "端口" .lang }}
                  type: integer
                  default: 80
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
      default: useExists
      description: {{ i18n "已存在的 CLB 实例：需先手工创建 CLB 实例，Ingress 实例删除时 CLB 实例不会被删除，重建 Ingress 实例后 VIP 不会发生变化<br>自动新建 CLB 实例：无需先手工创建 CLB 实例，Ingress 实例删除时 CLB 实例会随之删除，重建 Ingress 实例后 VIP 会发生变化；出于安全原因，自动新建 CLB 实例只支持内网，外网 CLB 请使用 “已存在的 CLB 实例”方案，内网子网 ID 请找资源管理员确认" .lang | quote }}
      ui:component:
        name: select
        props:
          clearable: false
          datasource:
            - label: {{ i18n "已存在的 CLB 实例" .lang }}
              value: useExists
            - label: {{ i18n "自动创建 CLB 实例" .lang }}
              value: autoCreate
              {{- if eq .clusterType "Shared" }}
              disabled: true
              tips: {{ i18n "共享集群暂不支持自动创建 CLB 实例" .lang }}
              {{- end }}
      ui:reactions:
        - target: spec.network.existLBID
          if: "{{`{{`}} $self.value === 'useExists' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
        - target: spec.network.subNetID
          if: "{{`{{`}} $self.value === 'autoCreate' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
    existLBID:
      title: "CLB ID"
      type: string
      default: lb-c5xxxxd6
      ui:component:
        props:
          placeholder: {{ i18n "例如：lb-c5xxxxd6" .lang | quote }}
      ui:rules:
        # 不是使用 "已存在 CLB 实例"时 或 非 qcloud 类型 ingress，可以不填该值
        - validator: "{{`{{`}} $self.getValue('controller.type') !== 'qcloud' || $self.getValue('spec.network.clbUseType') !== 'useExists' || $self.value {{`}}`}}"
          message: {{ i18n "值不能为空" .lang }}
    subNetID:
      title: {{ i18n "子网 ID" .lang }}
      type: string
      ui:component:
        props:
          placeholder: {{ i18n "例如：subnet-a3xxxxb4" .lang | quote }}
      ui:rules:
        # ingress 为 qcloud 类型 且 "自动创建 CLB 实例"，必须填值
        - validator: "{{`{{`}} $self.getValue('controller.type') !== 'qcloud' || $self.getValue('spec.network.clbUseType') !== 'autoCreate' || $self.value {{`}}`}}"
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
      description: {{ i18n "控制器类型 qcloud 暂时不支持配置默认后端" .lang }}
      default: ""
      ui:rules:
        - maxLength128
    port:
      title: {{ i18n "端口" .lang }}
      type: integer
      description: {{ i18n "控制器类型 qcloud 暂时不支持配置默认后端" .lang }}
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
    autoRewriteHttp:
      type: boolean
      title: {{ i18n "开启自动重定向" .lang }}
      description: {{ i18n "自动重定向：用户需要先创建出一个 HTTPS:443 监听器，并在其下创建转发规则。通过调用该接口，系统会自动创建出一个 HTTP:80 监听器（如果之前不存在），并在其下创建转发规则，与 HTTPS:443 监听器下的域名等各种配置对应。" .lang | quote }}
      ui:props:
        labelWidth: 350
    tls:
      type: array
      items:
        type: object
        properties:
          secretName:
            title: {{ i18n "证书" .lang }}
            type: string
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
      ui:reactions:
        - target: "spec.portConf.lb.useType"
          if: "{{`{{`}} $self.value === 'LoadBalancer' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
        - target: "spec.portConf.lb.existLBID"
          if: "{{`{{`}} $self.value === 'LoadBalancer' && $self.getValue('spec.portConf.lb.useType') === 'useExists' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
        - target: "spec.portConf.lb.subNetID"
          if: "{{`{{`}} $self.value === 'LoadBalancer' && $self.getValue('spec.portConf.lb.useType') === 'autoCreate' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
    lb:
      title: {{ i18n "负载均衡器" .lang }}
      type: object
      required:
        - existLBID
        - subNetID
      properties:
        useType:
          title: {{ i18n "CLB 使用方式" .lang }}
          type: string
          default: useExists
          description: {{ i18n "使用已有：仅支持使用当前未被 TKE 使用的应用型 CLB 以用于公网/内网访问 Service，请勿手动修改由 TKE 创建的 CLB 监听器<br>自动创建：自动创建内网 CLB 以提供内网访问入口，将提供一个可以被集群所有 VPC 下的其他资源访问的入口，支持 TCP/UDP 协议。需要被同一 VPC 下其他集群、云服务器等访问的服务可以选择 VPC 内网访问的形式，因安全原因，暂不支持公网 CLB 自动创建，内网子网 ID 请找资源管理员确认" .lang | quote }}
          ui:component:
            name: select
            props:
              datasource:
                - label: {{ i18n "使用已有" .lang }}
                  value: useExists
                - label: {{ i18n "自动创建" .lang }}
                  value: autoCreate
                  {{- if eq .clusterType "Shared" }}
                  disabled: true
                  tips: {{ i18n "共享集群暂不支持自动创建 CLB 实例" .lang }}
                  {{- end }}
          ui:reactions:
            - target: "spec.portConf.lb.existLBID"
              if: "{{`{{`}} $self.value === 'useExists' && $self.getValue('spec.portConf.type') == 'LoadBalancer' {{`}}`}}"
              then:
                state:
                  visible: true
              else:
                state:
                  visible: false
            - target: "spec.portConf.lb.subNetID"
              if: "{{`{{`}} $self.value === 'autoCreate' && $self.getValue('spec.portConf.type') == 'LoadBalancer' {{`}}`}}"
              then:
                state:
                  visible: true
              else:
                state:
                  visible: false
        existLBID:
          title: "CLB ID"
          type: string
          ui:component:
            props:
              placeholder: {{ i18n "例如：lb-c5xxxxd6" .lang | quote }}
          ui:rules:
            # 不是使用 "已有的 CLB 实例" 时 或 非 LoadBalancer 类型 Service，可以不填该值
            - validator: "{{`{{`}} $self.getValue('spec.portConf.type') !== 'LoadBalancer' || $self.getValue('spec.portConf.lb.useType') !== 'useExists' || $self.value {{`}}`}}"
              message: {{ i18n "值不能为空" .lang }}
        subNetID:
          title: {{ i18n "子网 ID" .lang }}
          type: string
          ui:component:
            props:
              placeholder: {{ i18n "例如：subnet-a3xxxxb4" .lang | quote }}
          ui:rules:
            # service 为 LoadBalancer 类型 且 "自动创建 CLB 实例" 时，必须填值
            - validator: "{{`{{`}} $self.getValue('spec.portConf.type') !== 'LoadBalancer' || $self.getValue('spec.portConf.lb.useType') !== 'autoCreate' || $self.value {{`}}`}}"
              message: {{ i18n "值不能为空" .lang }}
    ports:
      type: array
      minItems: 1
      items:
        type: object
        properties:
          name:
            title: {{ i18n "名称" .lang }}
            type: string
            default: "http"
            ui:rules:
              - maxLength64
              - rfc1123LabelRegex
            ui:props:
              showTitle: true
          port:
            title: {{ i18n "监听端口" .lang }}
            type: integer
            default: 80
            ui:component:
              props:
                min: 1
                max: 65535
            ui:rules:
              - validator: "{{`{{`}} $self.value {{`}}`}}"
                message: {{ i18n "值不能为空" .lang }}
            ui:props:
              showTitle: true                
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
            ui:props:
              showTitle: true                    
          targetPort:
            title: {{ i18n "目标端口" .lang }}
            type: string
            default: "80"
            ui:rules:
              - validator: "{{`{{`}} $self.value {{`}}`}}"
                message: {{ i18n "值不能为空" .lang }}
            ui:reactions:
              - if: "{{`{{`}} !$self.getValue('spec.selector.associate') {{`}}`}}"
                then:
                  state:
                    visible: true
                else:
                  state:
                    visible: false
                    value: "80"
            ui:props:
              showTitle: true                  
          targetSelectPort:
            title: {{ i18n "目标端口" .lang }}
            type: string
            default: "80"
            ui:rules:
              - validator: "{{`{{`}} $self.value {{`}}`}}"
                message: {{ i18n "值不能为空" .lang }}  
            ui:component:
              name: select
              props:
                remoteConfig:
                  url: "{{`{{`}} `${$context.baseUrl}/projects/${$context.projectID}/template/ports?kind=${$self.getValue('spec.selector.workloadType')}&templateSpace={{ .templateSpace }}&associateName=${$self.getValue('spec.selector.workloadName')}` {{`}}`}}"
            ui:reactions:
              - if: "{{`{{`}} $self.getValue('spec.selector.associate') {{`}}`}}"
                then:
                  state:
                    visible: true
                else:
                  state:
                    visible: false
                    value: "80"
              - lifetime: init
                then:
                  actions:
                    - "{{`{{`}} $loadDataSource {{`}}`}}"
              - source: "spec.selector.workloadName"
                then:
                  state:
                    value: ""
                  actions:
                    - "{{`{{`}} $loadDataSource {{`}}`}}"
            ui:props:
              showTitle: true                    
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
                    visible: false
                else:
                  state:
                    disabled: false
              - source: "spec.portConf.type"
                if: "{{`{{`}} $self.getValue('spec.portConf.type') === 'ClusterIP' {{`}}`}}"
                then:
                  state:
                    disabled: true
                    visible: false
                else:
                  state:
                    disabled: false
            ui:props:
              showTitle: true                    
      ui:component:
        name: bfArray
      ui:props:
        showTitle: false
{{- end }}

{{- define "network.svcSelector" }}
selector:
  title: {{ i18n "选择器" .lang }}
  type: object
  ui:order:
    - associate
    - workloadType
    - workloadName
    - labelSelected
    - labels
  properties:
    associate:
      title: {{ i18n "关联应用" .lang }}
      type: boolean
      default: false
      ui:reactions:
      - target: "{{`{{`}} $widgetNode?.getSibling('workloadType')?.id {{`}}`}}"
        if: "{{`{{`}} $self.value {{`}}`}}"
        then:
          state:
            visible: true
        else:
          state:
            visible: false
      - target: "{{`{{`}} $widgetNode?.getSibling('workloadName')?.id {{`}}`}}"
        if: "{{`{{`}} $self.value {{`}}`}}"
        then:
          state:
            visible: true
        else:
          state:
            visible: false
      - target: "{{`{{`}} $widgetNode?.getSibling('labelSelected')?.id {{`}}`}}"
        if: "{{`{{`}} $self.value {{`}}`}}"
        then:
          state:
            visible: true
        else:
          state:
            visible: false
      - target: "{{`{{`}} $widgetNode?.getSibling('labels')?.id {{`}}`}}"
        if: "{{`{{`}} $self.value {{`}}`}}"
        then:
          state:
            visible: false
        else:
          state:
            visible: true
      - target: spec.portConf.ports
        if: "{{`{{`}} $self.value {{`}}`}}"
        then:
          state:
            value: [{"name":"http","port":80,"protocol":"TCP","targetSelectPort":"", nodePort:0}] 
      - target: spec.portConf.ports
        if: "{{`{{`}} !$self.value {{`}}`}}"
        then:
          state:    
            value: [{"name":"http","port":80,"protocol":"TCP","targetPort":"", nodePort:0}]    
    workloadType:
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
    workloadName:
      title: {{ i18n "资源名称" .lang }}
      type: string
      default: deploy1
      ui:component:
        name: select
        props:
          remoteConfig:
            url: "{{`{{`}} `${$context.baseUrl}/projects/${$context.projectID}/template/resources?kind=${$self.getValue('spec.selector.workloadType')}&templateSpace={{ .templateSpace }}` {{`}}`}}"
      ui:reactions:
        - lifetime: init
          then:
            actions:
              - "{{`{{`}} $loadDataSource {{`}}`}}"
        - source: "spec.selector.workloadType"
          then:
            state:
              value: ""
            actions:
              - "{{`{{`}} $loadDataSource {{`}}`}}"
    labelSelected:
      title: {{ i18n "标签选择器" .lang }}
      type: object
      description: {{ i18n "若没有设置选择器，则不会自动创建 Endpoints，需要手动创建" .lang }}
      items:
        type: object
        properties:
          key:
            title: {{ i18n "键" .lang }}
            type: string
          value:
            title: {{ i18n "值" .lang }}
            type: string
      ui:component:
        name: kvSelector
        props:
          remoteConfig:
            url: "{{`{{`}} `${$context.baseUrl}/projects/${$context.projectID}/template/labels?kind=${$self.getValue('spec.selector.workloadType')}&templateSpace={{ .templateSpace }}&associateName=${$self.getValue('spec.selector.workloadName')}` {{`}}`}}"
      ui:reactions:
        - lifetime: init
          then:
            actions:
              - "{{`{{`}} $loadDataSource {{`}}`}}"
        - source: "spec.selector.workloadName"
          then:
            state:
              value: ""
            actions:
              - "{{`{{`}} $loadDataSource {{`}}`}}"
    labels:
      title: {{ i18n "标签选择器" .lang }}
      type: array
      description: {{ i18n "若没有设置选择器，则不会自动创建 Endpoints，需要手动创建" .lang }}
      items:
        type: object
        properties:
          key:
            title: {{ i18n "键" .lang }}
            type: string
            ui:rules:
              - required
              - labelKeyRegexWithVar
              - maxLength128
          value:
            title: {{ i18n "值" .lang }}
            type: string
            ui:rules:
              - maxLength128
      ui:component:
        name: bfArray
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
      title: {{ i18n "最大会话停留时间" .lang }}
      type: integer
      default: 10800
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
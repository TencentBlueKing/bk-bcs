{{- define "workload.deployReplicas" }}
replicas:
  title: 副本管理
  type: object
  properties:
    cnt:
      title: 副本数量
      type: integer
    updateStrategy:
      title: 升级策略
      type: string
      ui:component:
        name: radio
        props:
          datasource:
            - label: 滚动升级
              value: RollingUpdate
            - label: 重新创建
              value: Recreate
    maxSurge:
      title: 最大调度 Pod 数量
      type: integer
    msUnit:
      default: cnt
      title: 单位
      type: string
      ui:component:
        name: select
        props:
          clearable: false
          datasource:
            - label: '%'
              value: percent
            - label: 个
              value: cnt
    maxUnavailable:
      title: 最大不可用数量
      type: integer
    muaUnit:
      default: percent
      title: 单位
      type: string
      ui:component:
        name: select
        props:
          clearable: false
          datasource:
            - label: '%'
              value: percent
            - label: 个
              value: cnt
    minReadySecs:
      title: 最小就绪时间
      type: integer
      ui:component:
        name: unitInput
        props:
          unit: s
    progressDeadlineSecs:
      default: 0
      title: 进程截止时间
      type: integer
      ui:component:
        name: unitInput
        props:
          max: 86400
          unit: s
{{- end }}

{{- define "workload.nodeSelect" }}
nodeSelect:
  title: 节点选择
  type: object
  properties:
    nodeName:
      title: 节点名称
      type: string
      ui:component:
        name: select
        props:
          clearable: false
          datasource:
            - label: TODO
              value: TODO Get Node Data
    selector:
      title: 调度规则
      type: array
      items:
        properties:
          key:
            title: 键
            type: string
          value:
            title: 值
            type: string
        type: object
      ui:component:
        name: noTitleArray
    type:
      default: specificNode
      title: 节点类型
      type: string
      ui:component:
        name: radio
        props:
          datasource:
            - label: 任意可用节点
              value: anyAvailable
            - label: 指定节点
              value: specificNode
            - label: 调度规则匹配
              value: schedulingRule
      ui:reactions:
        - if: "{{`{{`}} $self.value === 'specificNode' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
          target: spec.nodeSelect.nodeName
        - if: "{{`{{`}} $self.value === 'schedulingRule' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
          target: spec.nodeSelect.selector
{{- end }}

{{- define "workload.affinity" }}
affinity:
  title: 亲和性/反亲和性
  type: object
  properties:
    {{- include "affinity.nodeAffinity" . | indent 4 }}
    {{- include "affinity.podAffinity" . | indent 4 }}
{{- end }}

{{- define "workload.toleration" }}
toleration:
  title: 污点/容忍
  type: object
  properties:
    rules:
      type: array
      items:
        type: object
        properties:
          effect:
            title: 影响
            type: string
            ui:component:
              name: select
              props:
                clearable: true
                datasource:
                  - label: 所有
                    value: All
                  - label: 不调度
                    value: NoSchedule
                  - label: 倾向不调度
                    value: PreferNoSchedule
                  - label: 不执行
                    value: NoExecute
          key:
            title: 键
            type: string
          op:
            title: 运算符
            type: string
            ui:component:
              name: select
              props:
                clearable: true
                datasource:
                  - label: Equal
                    value: Equal
                  - label: Exists
                    value: Exists
          tolerationSecs:
            default: 0
            title: 容忍时间
            type: integer
            ui:component:
              props:
                max: 86400
          value:
            title: 值
            type: string
      ui:component:
        name: noTitleArray
      ui:props:
        showTitle: false
{{- end }}

{{- define "workload.networking" }}
networking:
  title: 网络
  type: object
  properties:
    dnsPolicy:
      title: DNS 策略
      type: string
      ui:component:
        name: radio
        props:
          datasource:
            - label: Default
              value: Default
            - label: ClusterFirst
              value: ClusterFirst
            - label: None
              value: None
            - label: ClusterFirstWithHostNet
              value: ClusterFirstWithHostNet
    hostIPC:
      title: HostIPC
      type: boolean
    hostNetwork:
      title: HostNetwork
      type: boolean
    hostPID:
      title: PostPID
      type: boolean
    shareProcessNamespace:
      title: ShareProcessNamespace
      type: boolean
    hostName:
      title: 主机名称
      type: string
    subdomain:
      title: 域名
      type: string
    nameServers:
      title: 服务器地址
      type: array
      items:
        type: string
      ui:component:
        name: noTitleArray
    searches:
      title: 搜索域
      type: array
      items:
        type: string
      ui:component:
        name: noTitleArray
    dnsResolverOpts:
      title: DNS 解析
      type: array
      items:
        type: object
        properties:
          name:
            title: 键
            type: string
          value:
            title: 值
            type: string
      ui:component:
        name: noTitleArray
    hostAliases:
      title: 主机别名
      type: array
      items:
        type: object
        properties:
          alias:
            title: 主机别名
            type: string
          ip:
            title: IP 地址
            type: string
      ui:component:
        name: noTitleArray
{{- end }}

{{- define "workload.security" }}
security:
  title: 安全
  type: object
  properties:
    runAsUser:
      title: 用户
      type: integer
    runAsNonRoot:
      title: 以非 Root 运行
      type: boolean
    runAsGroup:
      title: 用户组
      type: integer
    fsGroup:
      type: integer
    seLinuxOpt:
      title: SELinux 选项
      type: object
      properties:
        level:
          type: string
        role:
          type: string
        type:
          type: string
        user:
          type: string
      ui:group:
        props:
          showTitle: true
{{- end }}

{{- define "workload.specOther" }}
other:
  title: 其他
  type: object
  properties:
    imagePullSecrets:
      title: 镜像拉取密钥
      type: array
      items:
        type: string
      ui:component:
        name: select
    restartPolicy:
      title: 重启策略
      type: string
      ui:component:
        name: radio
        props:
          datasource:
            - label: Always
              value: Always
            - label: OnFailure
              value: OnFailure
            - label: Never
              value: Never
    saName:
      title: 服务账号
      type: string
      ui:component:
        name: select
    terminationGracePeriodSecs:
      title: 终止容忍期
      type: integer
      ui:component:
        name: unitInput
        props:
          max: 86400
          unit: s
{{- end }}

{{- define "workload.volume" }}
volume:
  title: 数据卷
  type: object
  properties:
    {{- include "volume.pvc" . | indent 4 }}
    {{- include "volume.hostPath" . | indent 4 }}
    {{- include "volume.configMap" . | indent 4 }}
    {{- include "volume.secret" . | indent 4 }}
    {{- include "volume.emptyDir" . | indent 4 }}
    {{- include "volume.nfs" . | indent 4 }}
  ui:group:
    name: collapse
    props:
      border: true
      showTitle: true
      type: card
{{- end }}

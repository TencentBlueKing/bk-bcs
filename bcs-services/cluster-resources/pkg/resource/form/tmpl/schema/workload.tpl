{{- define "workload.deployReplicas" }}
replicas:
  title: 副本管理
  type: object
  properties:
    cnt:
      title: 副本数量
      type: integer
      default: 3
      ui:component:
        props:
          min: 1
    updateStrategy:
      title: 升级策略
      type: string
      default: RollingUpdate
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

{{- define "workload.dsReplicas" }}
replicas:
  title: 副本管理
  type: object
  properties:
    updateStrategy:
      title: 升级策略
      type: string
      default: RollingUpdate
      ui:component:
        name: radio
        props:
          datasource:
            - label: 滚动升级
              value: RollingUpdate
            - label: 重新创建
              value: Recreate
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
{{- end }}

{{- define "workload.stsReplicas" }}
replicas:
  title: 副本管理
  type: object
  properties:
    cnt:
      title: 副本数量
      type: integer
      default: 3
      ui:component:
        props:
          min: 1
    updateStrategy:
      title: 升级策略
      type: string
      default: RollingUpdate
      ui:component:
        name: radio
        props:
          datasource:
            - label: 滚动升级
              value: RollingUpdate
            - label: 重新创建
              value: Recreate
    podManPolicy:
      title: Pod 管理策略
      type: string
      default: OrderedReady
      ui:component:
        name: radio
        props:
          datasource:
            - label: OrderedReady
              value: OrderedReady
            - label: Parallel
              value: Parallel
{{- end }}

{{- define "workload.cjJobManage" }}
jobManage:
  title: 任务管理
  type: object
  required:
    - schedule
  properties:
    schedule:
      title: 调度规则
      type: string
      description: "CronTab 表达式，形如：*/10 * * * *"
      ui:rules:
        - required
        - maxLength64
    concurrencyPolicy:
      title: 并发策略
      type: string
      default: Allow
      ui:component:
        name: radio
        props:
          datasource:
            - label: 允许多个 Job 同时运行
              value: Allow
            - label: 若 Job 未结束，则跳过
              value: Forbid
            - label: 若 Job 未结束，则替换
              value: Replace
    suspend:
      title: 暂停
      type: bool
      default: false
    completions:
      title: 需完成数
      type: integer
      ui:component:
        props:
          max: 8192
    parallelism:
      title: 并发数
      type: integer
      ui:component:
        props:
          max: 256
    backoffLimit:
      title: 重试次数
      type: integer
      ui:component:
        props:
          max: 2048
    activeDDLSecs:
      title: 活跃终止时间
      type: integer
      ui:component:
        name: unitInput
        props:
          max: 1209600
          unit: s
    successfulJobsHistoryLimit:
      title: 历史累计成功数
      type: integer
      ui:component:
        props:
          max: 4096
    failedJobsHistoryLimit:
      title: 历史累计失败数
      type: integer
      ui:component:
        props:
          max: 4096
    startingDDLSecs:
      title: 运行截止时间
      type: integer
      ui:component:
        name: unitInput
        props:
          max: 1209600
          unit: s
{{- end }}

{{- define "workload.jobManage" }}
jobManage:
  title: 任务管理
  type: object
  properties:
    completions:
      title: 需完成数
      type: integer
      ui:component:
        props:
          max: 2048
    parallelism:
      title: 并发数
      type: integer
      ui:component:
        props:
          max: 256
    backoffLimit:
      title: 重试次数
      type: integer
      ui:component:
        props:
          max: 2048
    activeDDLSecs:
      title: 活跃终止时间
      type: integer
      ui:component:
        name: unitInput
        props:
          max: 1209600
          unit: s
{{- end }}

{{- define "workload.nodeSelect" }}
nodeSelect:
  title: 节点选择
  type: object
  required:
    - nodeName
    - selector
  properties:
    type:
      title: 节点类型
      type: string
      default: anyAvailable
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
    nodeName:
      title: 节点名称
      type: string
      ui:component:
        name: select
        props:
          clearable: false
          searchable: true
          remoteConfig:
            params:
              format: selectItems
            url: "{{`{{`}} `${$context.baseUrl}/projects/${$context.projectID}/clusters/${$context.clusterID}/nodes` {{`}}`}}"
      ui:reactions:
        - lifetime: init
          then:
            actions:
              - "{{`{{`}} $loadDataSource {{`}}`}}"
      ui:rules:
        - validator: "{{`{{`}} $self.getValue('spec.nodeSelect.type') !== 'specificNode' || ($self.getValue('spec.nodeSelect.type') === 'specificNode' && $self.value !== '') {{`}}`}}"
          message: "值不能为空"
    selector:
      title: 调度规则
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
            ui:rules:
              - maxLength128
        type: object
      ui:component:
        name: noTitleArray
      ui:rules:
        - validator: "{{`{{`}} $self.getValue('spec.nodeSelect.type') !== 'schedulingRule' || ($self.getValue('spec.nodeSelect.type') === 'schedulingRule' && $self.value.length > 0) {{`}}`}}"
          message: "至少包含一条调度规则"
  ui:order:
    - type
    - selector
    - nodeName
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
          key:
            title: 键
            type: string
            ui:rules:
              - required
              - maxLength128
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
          value:
            title: 值
            type: string
            ui:rules:
              - maxLength128
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
          tolerationSecs:
            default: 0
            title: 容忍时间
            type: integer
            ui:component:
              name: unitInput
              props:
                unit: s
                max: 86400
        ui:order:
          - key
          - op
          - value
          - effect
          - tolerationSecs
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
      default: ClusterFirst
      ui:component:
        name: radio
        props:
          datasource:
            - label: ClusterFirst
              value: ClusterFirst
            - label: ClusterFirstWithHostNet
              value: ClusterFirstWithHostNet
            - label: Default
              value: Default
            - label: None
              value: None
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
      ui:rules:
        - maxLength128
    subdomain:
      title: 域名
      type: string
      ui:rules:
        - maxLength128
    nameServers:
      title: 服务器地址
      type: array
      items:
        type: string
        ui:rules:
          - maxLength128
      ui:component:
        name: noTitleArray
    searches:
      title: 搜索域
      type: array
      items:
        type: string
        ui:rules:
          - maxLength128
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
            ui:rules:
              - maxLength128
          value:
            title: 值
            type: string
            ui:rules:
              - maxLength128
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
            ui:rules:
              - maxLength64
          ip:
            title: IP 地址
            type: string
            ui:rules:
              - maxLength64
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
          ui:rules:
            - maxLength64
        role:
          type: string
          ui:rules:
            - maxLength64
        type:
          type: string
          ui:rules:
            - maxLength64
        user:
          type: string
          ui:rules:
            - maxLength64
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
    restartPolicy:
      title: 重启策略
      type: string
      {{- if or (eq .kind "CronJob") (eq .kind "Job") }}
      default: OnFailure
      {{- else }}
      default: Always
      {{- end }}
      ui:component:
        name: radio
        props:
          datasource:
            {{- if and (ne .kind "CronJob") (ne .kind "Job") }}
            - label: Always
              value: Always
            {{- end }}
            - label: OnFailure
              value: OnFailure
            - label: Never
              value: Never
    saName:
      title: 服务账号
      type: string
      ui:component:
        name: select
        props:
          clearable: true
          searchable: true
          remoteConfig:
            params:
              format: selectItems
            url: "{{`{{`}} `${$context.baseUrl}/projects/${$context.projectID}/clusters/${$context.clusterID}/namespaces/${$self.getValue('metadata.namespace')}/rbac/service_accounts` {{`}}`}}"
      ui:reactions:
        - lifetime: init
          then:
            actions:
              - "{{`{{`}} $loadDataSource {{`}}`}}"
        - source: "metadata.namespace"
          then:
            actions:
              - "{{`{{`}} $loadDataSource {{`}}`}}"
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
  ui:order:
    - pvc
    - hostPath
    - configMap
    - secret
    - emptyDir
    - nfs
{{- end }}

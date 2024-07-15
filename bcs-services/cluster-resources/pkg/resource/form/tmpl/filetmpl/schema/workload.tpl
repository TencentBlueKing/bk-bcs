{{- define "workload.deployReplicas" }}
replicas:
  title: {{ i18n "副本管理" .lang }}
  type: object
  properties:
    cnt:
      title: {{ i18n "副本数量" .lang }}
      type: string
      default: "1"
    updateStrategy:
      title: {{ i18n "升级策略" .lang }}
      type: string
      default: RollingUpdate
      ui:component:
        name: radio
        props:
          datasource:
            - label: {{ i18n "滚动升级" .lang }}
              value: RollingUpdate
            - label: {{ i18n "重新创建" .lang }}
              value: Recreate
      ui:reactions:
        - target: "{{`{{`}} $widgetNode?.getSibling('maxSurge')?.id {{`}}`}}"
          if: "{{`{{`}} $self.value === 'RollingUpdate' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
        - target: "{{`{{`}} $widgetNode?.getSibling('msUnit')?.id {{`}}`}}"
          if: "{{`{{`}} $self.value === 'RollingUpdate' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
        - target: "{{`{{`}} $widgetNode?.getSibling('maxUnavailable')?.id {{`}}`}}"
          if: "{{`{{`}} $self.value === 'RollingUpdate' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
        - target: "{{`{{`}} $widgetNode?.getSibling('muaUnit')?.id {{`}}`}}"
          if: "{{`{{`}} $self.value === 'RollingUpdate' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
    maxSurge:
      title: {{ i18n "最大调度 Pod 数量" .lang }}
      type: integer
      default: 25
      ui:component:
        props:
          max: 4096
      ui:rules:
        - validator: "{{`{{`}} $self.getValue('spec.replicas.updateStrategy') === 'Recreate' || ($self.getValue('spec.replicas.maxUnavailable') !== 0 || $self.value !== 0) {{`}}`}}"
          message: {{ i18n "最大调度 Pod 数量 与最大不可用数量不可均为 0" .lang }}
    msUnit:
      title: {{ i18n "单位" .lang }}
      type: string
      default: percent
      ui:component:
        name: select
        props:
          clearable: false
          datasource:
            - label: '%'
              value: percent
            - label: {{ i18n "个" .lang }}
              value: cnt
    maxUnavailable:
      title: {{ i18n "最大不可用数量" .lang }}
      type: integer
      default: 25
      ui:component:
        props:
          max: 4096
      ui:rules:
        - validator: "{{`{{`}} $self.getValue('spec.replicas.updateStrategy') === 'Recreate' || (($self.getValue('spec.replicas.maxSurge') !== 0 || $self.value !== 0) && ($self.getValue('spec.replicas.muaUnit') !== 'percent' || $self.value <= 100)) {{`}}`}}"
          message: {{ i18n "最大调度 Pod 数量 与最大不可用数量不可均为 0，且最大不可用数量不可超过 100%" .lang }}
    muaUnit:
      title: {{ i18n "单位" .lang }}
      type: string
      default: percent
      ui:component:
        name: select
        props:
          clearable: false
          datasource:
            - label: '%'
              value: percent
            - label: {{ i18n "个" .lang }}
              value: cnt
      ui:reactions:
        - if: "{{`{{`}} $self.value === 'percent' {{`}}`}}"
          then:
            state:
              visible: true
              value: 25
          else:
            state:
              visible: true
              value: 1
          target: spec.replicas.maxUnavailable
    minReadySecs:
      title: {{ i18n "最小就绪时间" .lang }}
      type: integer
      default: 0
      ui:component:
        name: bfInput
        props:
          max: 2147483647
          unit: s
    progressDeadlineSecs:
      title: {{ i18n "进程截止时间" .lang }}
      type: integer
      default: 600
      ui:component:
        name: bfInput
        props:
          max: 2147483647
          unit: s
      ui:rules:
        - validator: "{{`{{`}} $self.getValue('spec.replicas.minReadySecs') < $self.value {{`}}`}}"
          message: {{ i18n "进程截止时间必须大于最小就绪时间" .lang }}
{{- end }}

{{- define "workload.dsReplicas" }}
replicas:
  title: {{ i18n "副本管理" .lang }}
  type: object
  properties:
    updateStrategy:
      title: {{ i18n "升级策略" .lang }}
      type: string
      default: RollingUpdate
      ui:component:
        name: radio
        props:
          datasource:
            - label: {{ i18n "滚动升级" .lang }}
              value: RollingUpdate
            - label: {{ i18n "手动删除" .lang }}
              value: OnDelete
      ui:reactions:
        - target: "{{`{{`}} $widgetNode?.getSibling('maxUnavailable')?.id {{`}}`}}"
          if: "{{`{{`}} $self.value === 'RollingUpdate' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
        - target: "{{`{{`}} $widgetNode?.getSibling('muaUnit')?.id {{`}}`}}"
          if: "{{`{{`}} $self.value === 'RollingUpdate' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
    maxUnavailable:
      title: {{ i18n "最大不可用数量" .lang }}
      type: integer
      default: 25
      ui:component:
        props:
          max: 4096
    muaUnit:
      title: {{ i18n "单位" .lang }}
      type: string
      default: percent
      ui:component:
        name: select
        props:
          clearable: false
          datasource:
            - label: '%'
              value: percent
            - label: {{ i18n "个" .lang }}
              value: cnt
      ui:reactions:
        - if: "{{`{{`}} $self.value === 'percent' {{`}}`}}"
          then:
            state:
              visible: true
              value: 25
          else:
            state:
              visible: true
              value: 1
          target: spec.replicas.maxUnavailable
    minReadySecs:
      title: {{ i18n "最小就绪时间" .lang }}
      type: integer
      default: 0
      ui:component:
        name: bfInput
        props:
          max: 86400
          unit: s
{{- end }}

{{- define "workload.stsReplicas" }}
replicas:
  title: {{ i18n "副本管理" .lang }}
  type: object
  properties:
    cnt:
      title: {{ i18n "副本数量" .lang }}
      type: string
      default: "1"
    svcName:
      title: {{ i18n "服务名称" .lang }}
      type: string
      ui:component:
        props:
          clearable: false
    updateStrategy:
      title: {{ i18n "升级策略" .lang }}
      type: string
      default: RollingUpdate
      ui:component:
        name: radio
        props:
          datasource:
            - label: {{ i18n "滚动升级" .lang }}
              value: RollingUpdate
            - label: {{ i18n "手动删除" .lang }}
              value: OnDelete
      ui:reactions:
        - target: spec.replicas.partition
          if: "{{`{{`}} $self.value === 'RollingUpdate' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
    podManPolicy:
      title: {{ i18n "Pod 管理策略" .lang }}
      type: string
      default: OrderedReady
      description: {{ i18n "OrderedReady：顺序启动或停止 Pod，若前一个 Pod 启动或停止未完成，则不会操作下一个 Pod<br>Parallel：启动或停止 Pod 前，不会检查 Pod 状态，直接进行并发操作<br>注意：此选项只会影响扩缩容操作，对更新操作无效！" .lang | quote }}
      ui:component:
        name: radio
        props:
          disabled: {{ eq .action "update" }}
          datasource:
            - label: OrderedReady
              value: OrderedReady
            - label: Parallel
              value: Parallel
    partition:
      title: {{ i18n "分区滚动更新" .lang }}
      type: integer
      default: 0
      description: {{ i18n "更新策略可以实现分区，所有序号大于该值的 Pod 都会被更新<br>例如：存在 Pod web-0 至 web-9，这里设置为 5，则只会更新 web-5 至 web-9，其他 Pod 不会被更新" .lang | quote }}
      ui:component:
        props:
          max: 8192
{{- end }}

{{- define "workload.stsVolumeClaimTmpl" }}
volumeClaimTmpl:
  title: {{ i18n "存储卷声明模板" .lang }}
  type: object
  properties:
    claims:
      title: {{ i18n "卷声明" .lang }}
      type: array
      items:
        type: object
        required:
          - pvcName
          - claimType
          - scName
          - pvName
          - storageSize
        properties:
          pvcName:
            title: {{ i18n "持久卷声明名称" .lang }}
            type: string
            ui:props:
              labelWidth: 350
            ui:rules:
              - required
              - maxLength128
              - nameRegexWithVar
          claimType:
            title: {{ i18n "卷声明类型" .lang }}
            type: string
            default: createBySC
            ui:component:
              name: radio
              props:
                datasource:
                  - label: {{ i18n "指定存储类以创建持久卷" .lang }}
                    value: createBySC
                  - label: {{ i18n "使用已存在的持久卷" .lang }}
                    value: useExistPV
            ui:reactions:
              - target: "{{`{{`}} $widgetNode?.getSibling('pvName')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'useExistPV' {{`}}`}}"
                then:
                  state:
                    visible: true
                else:
                  state:
                    visible: false
              - target: "{{`{{`}} $widgetNode?.getSibling('scName')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'createBySC' {{`}}`}}"
                then:
                  state:
                    visible: true
                else:
                  state:
                    visible: false
              - target: "{{`{{`}} $widgetNode?.getSibling('storageSize')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'createBySC' {{`}}`}}"
                then:
                  state:
                    visible: true
                else:
                  state:
                    visible: false
            ui:rules:
              - required
          pvName:
            title: {{ i18n "持久卷名称" .lang }}
            type: string
            ui:component:
              props:
                clearable: false
            ui:rules:
              - validator: "{{`{{`}} $widgetNode?.getSibling('claimType')?.instance?.value !== 'useExistPV' || $self.value !== '' {{`}}`}}"
                message: {{ i18n "值不能为空" .lang }}
          scName:
            title: {{ i18n "存储类名称" .lang }}
            type: string
            ui:component:
              props:
                clearable: false
            ui:rules:
              - validator: "{{`{{`}} $widgetNode?.getSibling('claimType')?.instance?.value !== 'createBySC' || $self.value !== '' {{`}}`}}"
                message: {{ i18n "值不能为空" .lang }}
          storageSize:
            title: {{ i18n "容量" .lang }}
            type: integer
            default: 10
            ui:component:
              name: bfInput
              props:
                max: 4096
                unit: Gi
            ui:rules:
              - validator: "{{`{{`}} $widgetNode?.getSibling('claimType')?.instance?.value !== 'createBySC' || $self.value !== 0 {{`}}`}}"
                message: {{ i18n "值不能为零" .lang }}
          accessModes:
            title: {{ i18n "访问模式" .lang }}
            type: array
            items:
              type: string
            ui:component:
              name: select
              props:
                clearable: true
                searchable: true
                datasource:
                  - label: ReadWriteOnce
                    value: RWO
                  - label: ReadOnlyMany
                    value: ROX
                  - label: ReadWriteMany
                    value: RWX
        ui:group:
          props:
            showTitle: false
            type: normal
          style:
            background: '#fff'
        ui:order:
          - pvcName
          - claimType
          - pvName
          - scName
          - storageSize
          - accessModes
      ui:group:
        props:
          showTitle: true
          type: card
        style:
          background: '#F5F7FA'
{{- end }}

{{- define "workload.cjJobManage" }}
jobManage:
  title: {{ i18n "任务管理" .lang }}
  type: object
  required:
    - schedule
  properties:
    schedule:
      title: {{ i18n "调度规则" .lang }}
      type: string
      description: {{ i18n "CronTab 表达式，形如：*/10 * * * *" .lang }}
      ui:rules:
        - required
        - maxLength64
    concurrencyPolicy:
      title: {{ i18n "并发策略" .lang }}
      type: string
      default: Allow
      ui:component:
        name: radio
        props:
          datasource:
            - label: {{ i18n "允许多个 Job 同时运行" .lang}}
              value: Allow
            - label: {{ i18n "若 Job 未结束，则跳过" .lang }}
              value: Forbid
            - label: {{ i18n "若 Job 未结束，则替换" .lang }}
              value: Replace
    suspend:
      title: {{ i18n "暂停" .lang }}
      type: boolean
      default: false
    completions:
      title: {{ i18n "需完成数" .lang }}
      type: integer
      ui:component:
        props:
          max: 8192
    parallelism:
      title: {{ i18n "并发数" .lang }}
      type: integer
      ui:component:
        props:
          max: 256
    backoffLimit:
      title: {{ i18n "重试次数" .lang }}
      type: integer
      ui:component:
        props:
          max: 2048
    activeDDLSecs:
      title: {{ i18n "活跃终止时间" .lang }}
      type: integer
      ui:component:
        name: bfInput
        props:
          max: 1209600
          unit: s
    successfulJobsHistoryLimit:
      title: {{ i18n "历史累计成功数" .lang }}
      type: integer
      ui:component:
        props:
          max: 4096
    failedJobsHistoryLimit:
      title: {{ i18n "历史累计失败数" .lang }}
      type: integer
      ui:component:
        props:
          max: 4096
    startingDDLSecs:
      title: {{ i18n "运行截止时间" .lang }}
      type: integer
      ui:component:
        name: bfInput
        props:
          max: 1209600
          unit: s
{{- end }}

{{- define "workload.jobManage" }}
jobManage:
  title: {{ i18n "任务管理" .lang }}
  type: object
  properties:
    completions:
      title: {{ i18n "需完成数" .lang }}
      type: integer
      ui:component:
        props:
          max: 2048
    parallelism:
      title: {{ i18n "并发数" .lang }}
      type: integer
      ui:component:
        props:
          max: 256
    backoffLimit:
      title: {{ i18n "重试次数" .lang }}
      type: integer
      ui:component:
        props:
          max: 2048
    activeDDLSecs:
      title: {{ i18n "活跃终止时间" .lang }}
      type: integer
      ui:component:
        name: bfInput
        props:
          max: 1209600
          unit: s
{{- end }}

{{- define "workload.nodeSelect" }}
nodeSelect:
  title: {{ i18n "节点选择" .lang }}
  type: object
  required:
    - nodeName
    - selector
  properties:
    type:
      title: {{ i18n "节点类型" .lang }}
      type: string
      default: anyAvailable
      ui:component:
        name: radio
        props:
          datasource:
            - label: {{ i18n "任意可用节点" .lang }}
              value: anyAvailable
            - label: {{ i18n "指定节点" .lang }}
              value: specificNode
            - label: {{ i18n "调度规则匹配" .lang }}
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
              value: []
              visible: false
          target: spec.nodeSelect.selector
    nodeName:
      title: {{ i18n "节点名称" .lang }}
      type: string
      ui:component:
        props:
          clearable: false
      ui:rules:
        - validator: "{{`{{`}} $self.getValue('spec.nodeSelect.type') !== 'specificNode' || ($self.getValue('spec.nodeSelect.type') === 'specificNode' && $self.value !== '') {{`}}`}}"
          message: {{ i18n "值不能为空" .lang }}
    selector:
      title: {{ i18n "调度规则" .lang }}
      type: array
      items:
        properties:
          key:
            title: {{ i18n "键" .lang }}
            type: string
            ui:rules:
              - required
              - maxLength128
          value:
            title: {{ i18n "值" .lang }}
            type: string
            ui:rules:
              - maxLength128
        type: object
      ui:component:
        name: bfArray
      ui:rules:
        - validator: "{{`{{`}} $self.getValue('spec.nodeSelect.type') !== 'schedulingRule' || ($self.getValue('spec.nodeSelect.type') === 'schedulingRule' && $self.value.length > 0) {{`}}`}}"
          message: {{ i18n "至少包含一条调度规则" .lang }}
  ui:order:
    - type
    - selector
    - nodeName
{{- end }}


{{- define "workload.labels" }}
labels:
  title: {{ i18n "标签管理" .lang }}
  type: object
  properties:
    {{- if (hasLabelSelector .kind) }}
    labels:
      title: {{ i18n "选择器" .lang }}
      type: array
      default: {{ .selectorLabel }}
      description: {{ i18n "标签选择器在创建资源后是不可变的，请务必小心谨慎更改选择器。" .lang }}
      ui:rules:
        - sliceLength1
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
    {{- end }}
    templateLabels:
      title: Pod {{ i18n "标签" .lang }}
      type: array
      default: {{ .selectorLabel }}
      ui:rules:
        - sliceLength1
      description: {{ i18n "该标签将添加到 Pod 标签中。" .lang }}
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
    {{- if eq .kind "CronJob" }}
    jobTemplatelabels:
      title: Job {{ i18n "标签" .lang }}
      type: array
      description: {{ i18n "该标签将添加到 Job 标签中。" .lang }}
      items:
        properties:
          key:
            title: {{ i18n "键" .lang }}
            type: string
            ui:rules:
              - required
              - maxLength128
              - labelKeyRegexWithVar
          value:
            title: {{ i18n "值" .lang }}
            type: string
            ui:rules:
              - maxLength64
              - labelValRegexWithVar
        type: object
      ui:component:
        name: bfArray
    {{- end }}
  ui:order:
    - labels
    - templateLabels
    - jobTemplatelabels
{{- end }}

{{- define "workload.affinity" }}
affinity:
  title: {{ i18n "亲和性/反亲和性" .lang }}
  type: object
  properties:
    {{- include "affinity.nodeAffinity" . | indent 4 }}
    {{- include "affinity.podAffinity" . | indent 4 }}
{{- end }}

{{- define "workload.toleration" }}
toleration:
  title: {{ i18n "污点/容忍" .lang }}
  type: object
  properties:
    rules:
      type: array
      items:
        type: object
        properties:
          key:
            title: {{ i18n "键" .lang }}
            type: string
            ui:rules:
              - validator: "{{`{{`}} $widgetNode.getSibling('op').value === 'Exists' || $self.value !== '' {{`}}`}}"
                message: {{ i18n "键不能为空" .lang }}
              - maxLength128
          op:
            title: {{ i18n "运算符" .lang }}
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
            ui:reactions:
              - target: "{{`{{`}} $widgetNode?.getSibling('value')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'Exists' {{`}}`}}"
                then:
                  state:
                    disabled: true
                    value: ""
                else:
                  state:
                    disabled: false
          value:
            title: {{ i18n "值" .lang }}
            type: string
            ui:rules:
              - maxLength128
          effect:
            title: {{ i18n "影响" .lang }}
            type: string
            default: NoSchedule
            ui:rules:
              - required
            ui:component:
              name: select
              props:
                clearable: true
                datasource:
                  - label: {{ i18n "不调度（NoSchedule）" .lang }}
                    value: NoSchedule
                  - label: {{ i18n "倾向不调度（PreferNoSchedule）" .lang }}
                    value: PreferNoSchedule
                  - label: {{ i18n "不执行（NoExecute）" .lang }}
                    value: NoExecute
            ui:reactions:
              - target: "{{`{{`}} $widgetNode?.getSibling('tolerationSecs')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'NoExecute' {{`}}`}}"
                then:
                  state:
                    disabled: false
                else:
                  state:
                    disabled: true
                    value: 0
          tolerationSecs:
            default: 0
            title: {{ i18n "容忍时间" .lang }}
            type: integer
            ui:component:
              name: bfInput
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
        name: bfArray
      ui:props:
        showTitle: false
{{- end }}

{{- define "workload.networking" }}
networking:
  title: {{ i18n "网络" .lang }}
  type: object
  properties:
    dnsPolicy:
      title: {{ i18n "DNS 策略" .lang }}
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
      title: {{ i18n "主机 IPC" .lang }}
      type: boolean
    hostNetwork:
      title: {{ i18n "主机网络" .lang }}
      type: boolean
    hostPID:
      title: {{ i18n "主机 PID" .lang }}
      type: boolean
      description: {{ i18n "主机 PID 与共享进程命名空间不可同时启用" .lang }}
      ui:reactions:
        - target: "{{`{{`}} $widgetNode?.getSibling('shareProcessNamespace')?.id {{`}}`}}"
          if: "{{`{{`}} $self.value === true {{`}}`}}"
          then:
            state:
              value: false
    shareProcessNamespace:
      title: {{ i18n "共享进程命名空间" .lang }}
      type: boolean
      description: {{ i18n "主机 PID 与共享进程命名空间不可同时启用" .lang }}
      ui:reactions:
        - target: "{{`{{`}} $widgetNode?.getSibling('hostPID')?.id {{`}}`}}"
          if: "{{`{{`}} $self.value === true {{`}}`}}"
          then:
            state:
              value: false
    hostname:
      title: {{ i18n "主机名称" .lang }}
      type: string
      ui:rules:
        - maxLength128
    subdomain:
      title: {{ i18n "子域名" .lang }}
      type: string
      ui:rules:
        - maxLength128
    nameServers:
      title: {{ i18n "服务器地址" .lang }}
      type: array
      items:
        type: string
        ui:rules:
          - maxLength128
      ui:component:
        name: bfArray
      ui:rules:
        - validator: "{{`{{`}} $self.getValue('spec.networking.dnsPolicy') !== 'None' || $self.value.length > 0 {{`}}`}}"
          message: {{ i18n "至少包含一个服务器地址" .lang }}
    searches:
      title: {{ i18n "搜索域" .lang }}
      type: array
      items:
        type: string
        ui:rules:
          - maxLength128
      ui:component:
        name: bfArray
    dnsResolverOpts:
      title: {{ i18n "DNS 解析" .lang }}
      type: array
      items:
        type: object
        properties:
          name:
            title: {{ i18n "键" .lang }}
            type: string
            ui:rules:
              - maxLength128
          value:
            title: {{ i18n "值" .lang }}
            type: string
            ui:rules:
              - maxLength128
      ui:component:
        name: bfArray
    hostAliases:
      title: {{ i18n "主机别名" .lang }}
      type: array
      items:
        type: object
        properties:
          alias:
            title: {{ i18n "别名" .lang }}
            type: string
            ui:component:
              props:
                placeholder: {{ i18n "别名（多个值请以英文逗号分隔）" .lang }}
            ui:rules:
              - maxLength250
          ip:
            title: {{ i18n "IP 地址" .lang }}
            type: string
            ui:rules:
              - maxLength64
      ui:component:
        name: bfArray
{{- end }}

{{- define "workload.security" }}
security:
  title: {{ i18n "安全" .lang }}
  type: object
  properties:
    runAsUser:
      title: {{ i18n "用户" .lang }}
      type: integer
      ui:component:
        props:
          max: 65535
    runAsNonRoot:
      title: {{ i18n "以非 Root 运行" .lang }}
      type: boolean
    runAsGroup:
      title: {{ i18n "用户组" .lang }}
      type: integer
      ui:component:
        props:
          max: 65535
    fsGroup:
      type: integer
      ui:component:
        props:
          max: 65535
    seLinuxOpt:
      title: {{ i18n "SELinux 选项" .lang }}
      type: object
      properties:
        level:
          type: string
          title: Level
          ui:rules:
            - maxLength64
        role:
          type: string
          title: Role
          ui:rules:
            - maxLength64
        type:
          type: string
          title: Type
          ui:rules:
            - maxLength64
        user:
          type: string
          title: User
          ui:rules:
            - maxLength64
      ui:group:
        props:
          showTitle: true
{{- end }}

{{- define "workload.readinessGates" }}
readinessGates:
  title: Readiness Gates
  type: object
  properties:
    readinessGates:
      title: Readiness Gates
      type: array
      items:
        type: string
        title: {{ i18n "条件类型" .lang }}
        ui:rules:
          - maxLength128
          - labelKeyRegexWithVar
      ui:component:
        name: bfArray
{{- end }}

{{- define "workload.specOther" }}
other:
  title: {{ i18n "其他" .lang }}
  type: object
  properties:
    imagePullSecrets:
      title: {{ i18n "镜像拉取密钥" .lang }}
      type: array
      items:
        type: string
      ui:component:
        props:
          clearable: true
    restartPolicy:
      title: {{ i18n "重启策略" .lang }}
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
            # CJ, Job 类型只会有 OnFailure，Never
            {{- if and (ne .kind "CronJob") (ne .kind "Job") }}
            - label: Always
              value: Always
            {{- end }}
            # Deploy, DS, STS 只会有 Always
            {{- if and (ne .kind "Deployment") (ne .kind "DaemonSet") (ne .kind "StatefulSet") }}
            - label: OnFailure
              value: OnFailure
            - label: Never
              value: Never
            {{- end }}
    saName:
      title: {{ i18n "服务账号" .lang }}
      type: string
      ui:component:
        props:
          clearable: true
    terminationGracePeriodSecs:
      title: {{ i18n "终止容忍期" .lang }}
      type: integer
      ui:component:
        name: bfInput
        props:
          max: 86400
          unit: s
{{- end }}

{{- define "workload.volume" }}
volume:
  title: {{ i18n "数据卷" .lang }}
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
      border: false
      showTitle: false
      verifiable: true
  ui:order:
    - pvc
    - hostPath
    - configMap
    - secret
    - emptyDir
    - nfs
{{- end }}

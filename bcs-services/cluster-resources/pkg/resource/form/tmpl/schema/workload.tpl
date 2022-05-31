{{- define "workload.deployReplicas" }}
replicas:
  title: {{ i18n "副本管理" .lang }}
  type: object
  properties:
    cnt:
      title: {{ i18n "副本数量" .lang }}
      type: integer
      default: 3
      ui:component:
        props:
          max: 4096
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
    maxSurge:
      title: {{ i18n "最大调度 Pod 数量" .lang }}
      type: integer
      ui:component:
        props:
          max: 4096
    msUnit:
      title: {{ i18n "单位" .lang }}
      type: string
      default: cnt
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
      ui:component:
        props:
          max: 4096
    muaUnit:
      title: {{ i18n "单位" .lang }}
      type: string
      default: cnt
      ui:component:
        name: select
        props:
          clearable: false
          datasource:
            - label: '%'
              value: percent
            - label: {{ i18n "个" .lang }}
              value: cnt
    minReadySecs:
      title: {{ i18n "最小就绪时间" .lang }}
      type: integer
      ui:component:
        name: unitInput
        props:
          max: 2147483647
          unit: s
    progressDeadlineSecs:
      default: 0
      title: {{ i18n "进程截止时间" .lang }}
      type: integer
      ui:component:
        name: unitInput
        props:
          max: 2147483647
          unit: s
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
            - label: {{ i18n "重新创建" .lang }}
              value: Recreate
    maxUnavailable:
      title: {{ i18n "最大不可用数量" .lang }}
      type: integer
      ui:component:
        props:
          max: 4096
    muaUnit:
      default: percent
      title: {{ i18n "单位" .lang }}
      type: string
      ui:component:
        name: select
        props:
          clearable: false
          datasource:
            - label: '%'
              value: percent
            - label: {{ i18n "个" .lang }}
              value: cnt
    minReadySecs:
      title: {{ i18n "最小就绪时间" .lang }}
      type: integer
      ui:component:
        name: unitInput
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
      type: integer
      default: 3
      ui:component:
        props:
          max: 4096
    svcName:
      title: {{ i18n "服务名称" .lang }}
      type: string
      ui:component:
        name: select
        props:
          clearable: false
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
            actions:
              - "{{`{{`}} $loadDataSource {{`}}`}}"
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
    podManPolicy:
      title: {{ i18n "Pod 管理策略" .lang }}
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
              - nameRegex
          claimType:
            title: {{ i18n "卷声明类型" .lang }}
            type: string
            default: useExistPV
            ui:component:
              name: radio
              props:
                datasource:
                  - label: {{ i18n "使用已存在的持久卷" .lang }}
                    value: useExistPV
                  - label: {{ i18n "指定存储类以创建持久卷" .lang }}
                    value: createBySC
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
              name: select
              props:
                clearable: false
                searchable: true
                remoteConfig:
                  params:
                    format: selectItems
                  url: "{{`{{`}} `${$context.baseUrl}/projects/${$context.projectID}/clusters/${$context.clusterID}/storages/persistent_volumes` {{`}}`}}"
            ui:reactions:
              - lifetime: init
                then:
                  actions:
                    - "{{`{{`}} $loadDataSource {{`}}`}}"
            # ui:rules:
            # TODO claimType == useExistPV 必填
          scName:
            title: {{ i18n "存储类名称" .lang }}
            type: string
            ui:component:
              name: select
              props:
                clearable: false
                searchable: true
                remoteConfig:
                  params:
                    format: selectItems
                  url: "{{`{{`}} `${$context.baseUrl}/projects/${$context.projectID}/clusters/${$context.clusterID}/storages/storage_classes` {{`}}`}}"
            ui:reactions:
              - lifetime: init
                then:
                  actions:
                    - "{{`{{`}} $loadDataSource {{`}}`}}"
            # ui:rules:
            # TODO claimType == createBySC 必填
          storageSize:
            title: {{ i18n "容量" .lang }}
            type: integer
            default: 10
            ui:component:
              name: unitInput
              props:
                max: 4096
                unit: Gi
            # ui:rules:
            # TODO claimType == createBySC 必填
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
      type: bool
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
        name: unitInput
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
        name: unitInput
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
        name: unitInput
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
              visible: false
          target: spec.nodeSelect.selector
    nodeName:
      title: {{ i18n "节点名称" .lang }}
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
        name: noTitleArray
      ui:rules:
        - validator: "{{`{{`}} $self.getValue('spec.nodeSelect.type') !== 'schedulingRule' || ($self.getValue('spec.nodeSelect.type') === 'schedulingRule' && $self.value.length > 0) {{`}}`}}"
          message: {{ i18n "至少包含一条调度规则" .lang }}
  ui:order:
    - type
    - selector
    - nodeName
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
              - required
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
            ui:component:
              name: select
              props:
                clearable: true
                datasource:
                  - label: {{ i18n "所有" .lang }}
                    value: All
                  - label: {{ i18n "不调度" .lang }}
                    value: NoSchedule
                  - label: {{ i18n "倾向不调度" .lang }}
                    value: PreferNoSchedule
                  - label: {{ i18n "不执行" .lang }}
                    value: NoExecute
          tolerationSecs:
            default: 0
            title: {{ i18n "容忍时间" .lang }}
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
    shareProcessNamespace:
      title: {{ i18n "共享进程命名空间" .lang }}
      type: boolean
    hostName:
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
        name: noTitleArray
    searches:
      title: {{ i18n "搜索域" .lang }}
      type: array
      items:
        type: string
        ui:rules:
          - maxLength128
      ui:component:
        name: noTitleArray
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
        name: noTitleArray
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
        name: noTitleArray
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
            {{- if and (ne .kind "CronJob") (ne .kind "Job") }}
            - label: Always
              value: Always
            {{- end }}
            - label: OnFailure
              value: OnFailure
            - label: Never
              value: Never
    saName:
      title: {{ i18n "服务账号" .lang }}
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
      title: {{ i18n "终止容忍期" .lang }}
      type: integer
      ui:component:
        name: unitInput
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

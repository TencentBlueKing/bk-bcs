{{- define "container.containerGroup" }}
containerGroup:
  title: {{ i18n "容器组" .lang }}
  type: object
  properties:
    {{- include "container.initContainers" . | indent 4 }}
    {{- include "container.containers" . | indent 4 }}
  ui:group:
    name: collapse
    props:
      border: true
      showTitle: true
      type: card
      verifiable: true
      defaultActiveName:
        - containers
  ui:order:
    - initContainers
    - containers
{{- end }}

{{- define "container.initContainers" }}
initContainers:
  title: {{ i18n "初始容器" .lang }}
  type: array
  items:
    type: object
    properties:
      {{- include "container.basic" (dict "lang" .lang "defaultImage" "busybox:latest" "defaultName" "init") | indent 6 }}
      {{- include "container.command" . | indent 6 }}
      {{- include "container.service" . | indent 6 }}
      {{- include "container.envs" . | indent 6 }}
      {{- include "container.resource" . | indent 6 }}
      {{- include "container.security" . | indent 6 }}
      {{- include "container.mount" . | indent 6 }}
    ui:group:
      name: tab
      style:
        background: '#fff'
    ui:order:
      - basic
      - command
      - service
      - envs
      - resource
      - security
      - mount
  ui:group:
    props:
      showTitle: false
{{- end }}

{{- define "container.containers" }}
containers:
  title: {{ i18n "标准容器" .lang }}
  type: array
  minItems: 1
  items:
    type: object
    properties:
      # 标准容器的 默认镜像 和 容器名称 与初始容器不同
      {{- include "container.basic" (dict "lang" .lang "defaultImage" "nginx:latest" "defaultName" "main")  | indent 6 }}
      {{- include "container.command" . | indent 6 }}
      {{- include "container.service" . | indent 6 }}
      {{- include "container.envs" . | indent 6 }}
      {{- include "container.healthz" . | indent 6 }}
      {{- include "container.resource" . | indent 6 }}
      {{- include "container.security" . | indent 6 }}
      {{- include "container.mount" . | indent 6 }}
    ui:group:
      name: tab
      style:
        background: '#fff'
    ui:order:
      - basic
      - command
      - service
      - envs
      - healthz
      - resource
      - security
      - mount
  ui:group:
    props:
      showTitle: false
{{- end }}

{{- define "container.basic" }}
basic:
  title: {{ i18n "基础信息" .lang }}
  type: object
  required:
    - name
    - image
  properties:
    name:
      title: {{ i18n "容器名称" .lang }}
      type: string
      default: {{ .defaultName }}
      ui:rules:
        - required
        - maxLength64
    image:
      title: {{ i18n "容器镜像" .lang }}
      type: string
      default: {{ .defaultImage }}
      ui:rules:
        - required
        - maxLength128
    pullPolicy:
      title: {{ i18n "拉取策略" .lang }}
      type: string
      default: IfNotPresent
      ui:component:
        name: select
        props:
          clearable: false
          datasource:
            - label: IfNotPresent
              value: IfNotPresent
            - label: Always
              value: Always
            - label: Never
              value: Never
{{- end }}

{{- define "container.command" }}
command:
  title: {{ i18n "命令" .lang }}
  type: object
  properties:
    workingDir:
      title: {{ i18n "工作目录" .lang }}
      type: string
      ui:rules:
        - maxLength128
    stdin:
      title: {{ i18n "标准输入" .lang }}
      type: boolean
    stdinOnce:
      title: {{ i18n "仅一次" .lang }}
      type: boolean
    tty:
      title: tty
      type: boolean
    command:
      title: {{ i18n "命令" .lang }}
      type: array
      items:
        type: string
        ui:rules:
          - maxLength250
      ui:component:
        name: bfArray
    args:
      title: {{ i18n "参数" .lang }}
      type: array
      items:
        type: string
        ui:rules:
          - maxLength250
      ui:component:
        name: bfArray
{{- end }}

{{- define "container.service" }}
service:
  title: {{ i18n "服务端口" .lang }}
  type: object
  properties:
    ports:
      type: array
      items:
        type: object
        properties:
          name:
            title: {{ i18n "名称" .lang }}
            type: string
            ui:rules:
              - required
              - maxLength64
          containerPort:
            title: {{ i18n "容器端口" .lang }}
            type: integer
            ui:component:
              props:
                max: 65535
          protocol:
            title: {{ i18n "协议" .lang }}
            type: string
            default: TCP
            ui:component:
              name: select
              props:
                datasource:
                  - label: TCP
                    value: TCP
                  - label: UDP
                    value: UDP
          hostPort:
            title: {{ i18n "主机端口" .lang }}
            type: integer
            ui:component:
              props:
                max: 65535
      ui:component:
        name: bfArray
      ui:props:
        showTitle: false
{{- end }}

{{- define "container.envs" }}
envs:
  title: {{ i18n "环境变量" .lang }}
  type: object
  properties:
    vars:
      type: array
      items:
        type: object
        properties:
          type:
            title: {{ i18n "类型" .lang }}
            type: string
            default: keyValue
            ui:component:
              name: select
              props:
                clearable: false
                datasource:
                  - label: Key-Value
                    value: keyValue
                  - label: Pod Field
                    value: podField
                  - label: Resource
                    value: resource
                  - label: ConfigMap Key
                    value: configMapKey
                  - label: Secret Key
                    value: secretKey
                  - label: ConfigMap
                    value: configMap
                  - label: Secret
                    value: secret
            ui:reactions:
              - target: "{{`{{`}} $widgetNode?.getSibling('source')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'keyValue' || $self.value === 'podField' {{`}}`}}"
                then:
                  state:
                    disabled: true
                else:
                  state:
                    disabled: false
              - target: "{{`{{`}} $widgetNode?.getSibling('value')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'configMap' || $self.value === 'secret' {{`}}`}}"
                then:
                  state:
                    disabled: true
                else:
                  state:
                    disabled: false
            ui:rules:
              - required
          name:
            title: {{ i18n "内容（Name/Prefix）" .lang }}
            type: string
            ui:rules:
              - required
              - maxLength128
          source:
            title: {{ i18n "来源" .lang }}
            type: string
            ui:rules:
              - maxLength128
              - validator: "{{`{{`}} ($widgetNode?.getSibling('type')?.instance?.value === 'keyValue' || $widgetNode?.getSibling('type')?.instance?.value === 'podField') || $self.value !== '' {{`}}`}}"
                message: {{ i18n "值不能为空" .lang }}
          value:
            title: {{ i18n "值" .lang }}
            type: string
            ui:rules:
              - maxLength128
              - validator: "{{`{{`}} ($widgetNode?.getSibling('type')?.instance?.value === 'keyValue' || $widgetNode?.getSibling('type')?.instance?.value === 'configMap' || $widgetNode?.getSibling('type')?.instance?.value === 'secret') || $self.value !== '' {{`}}`}}"
                message: {{ i18n "值不能为空" .lang }}
      ui:component:
        name: bfArray
      ui:props:
        showTitle: false
{{- end }}

{{- define "container.healthz" }}
healthz:
  title: {{ i18n "健康检查" .lang }}
  type: object
  properties:
    readinessProbe:
      title: {{ i18n "就绪探针" .lang }}
      type: object
      {{- include "container.probe" . | indent 6 }}
    livenessProbe:
      title: {{ i18n "存活探针" .lang }}
      type: object
      {{- include "container.probe" . | indent 6 }}
  ui:group:
    name: collapse
  ui:order:
    - readinessProbe
    - livenessProbe
{{- end }}

{{- define "container.probe" }}
properties:
  enabled:
    title: {{ i18n "启用" .lang }}
    type: boolean
    default: false
    ui:reactions:
      - target: "{{`{{`}} $widgetNode?.getSibling('type')?.id {{`}}`}}"
        if: "{{`{{`}} $self.value {{`}}`}}"
        then:
          state:
            visible: true
        else:
          state:
            visible: false
      - target: "{{`{{`}} $widgetNode?.getSibling('port')?.id {{`}}`}}"
        if: "{{`{{`}} $self.value && $widgetNode?.getSibling('type')?.instance?.value !== 'exec' {{`}}`}}"
        then:
          state:
            visible: true
        else:
          state:
            value: 0
            visible: false
      - target: "{{`{{`}} $widgetNode?.getSibling('path')?.id {{`}}`}}"
        if: "{{`{{`}} $self.value && $widgetNode?.getSibling('type')?.instance?.value == 'httpGet' {{`}}`}}"
        then:
          state:
            visible: true
        else:
          state:
            value: ""
            visible: false
      - target: "{{`{{`}} $widgetNode?.getSibling('command')?.id {{`}}`}}"
        if: "{{`{{`}} $self.value && $widgetNode?.getSibling('type')?.instance?.value === 'exec' {{`}}`}}"
        then:
          state:
            visible: true
        else:
          state:
            value: []
            visible: false
      - target: "{{`{{`}} $widgetNode?.getSibling('initialDelaySecs')?.id {{`}}`}}"
        if: "{{`{{`}} $self.value {{`}}`}}"
        then:
          state:
            visible: true
        else:
          state:
            visible: false
      - target: "{{`{{`}} $widgetNode?.getSibling('periodSecs')?.id {{`}}`}}"
        if: "{{`{{`}} $self.value {{`}}`}}"
        then:
          state:
            visible: true
        else:
          state:
            visible: false
      - target: "{{`{{`}} $widgetNode?.getSibling('timeoutSecs')?.id {{`}}`}}"
        if: "{{`{{`}} $self.value {{`}}`}}"
        then:
          state:
            visible: true
        else:
          state:
            visible: false
      - target: "{{`{{`}} $widgetNode?.getSibling('successThreshold')?.id {{`}}`}}"
        if: "{{`{{`}} $self.value {{`}}`}}"
        then:
          state:
            visible: true
        else:
          state:
            visible: false
      - target: "{{`{{`}} $widgetNode?.getSibling('failureThreshold')?.id {{`}}`}}"
        if: "{{`{{`}} $self.value {{`}}`}}"
        then:
          state:
            visible: true
        else:
          state:
            visible: false
  type:
    title: {{ i18n "检查类型" .lang }}
    type: string
    default: httpGet
    ui:component:
      name: select
      props:
        clearable: false
        datasource:
          - label: httpGet
            value: httpGet
          - label: tcpSocket
            value: tcpSocket
          - label: exec
            value: exec
    ui:reactions:
      - target: "{{`{{`}} $widgetNode?.getSibling('port')?.id {{`}}`}}"
        if: "{{`{{`}} !$widgetNode?.getSibling('enabled')?.instance?.value || $self.value === 'exec' {{`}}`}}"
        then:
          state:
            value: 0
            visible: false
        else:
          state:
            visible: true
      - target: "{{`{{`}} $widgetNode?.getSibling('path')?.id {{`}}`}}"
        if: "{{`{{`}} $widgetNode?.getSibling('enabled')?.instance?.value && $self.value === 'httpGet' {{`}}`}}"
        then:
          state:
            visible: true
        else:
          state:
            value: ""
            visible: false
      - target: "{{`{{`}} $widgetNode?.getSibling('command')?.id {{`}}`}}"
        if: "{{`{{`}} $widgetNode?.getSibling('enabled')?.instance?.value && $self.value === 'exec' {{`}}`}}"
        then:
          state:
            visible: true
        else:
          state:
            value: []
            visible: false
  port:
    title: {{ i18n "端口" .lang }}
    type: integer
    ui:component:
      props:
        max: 65535
    ui:rules:
      - validator: "{{`{{`}} !$widgetNode?.getSibling('enabled')?.instance?.value || ($widgetNode?.getSibling('type')?.instance?.value !== 'httpGet' && $widgetNode?.getSibling('type')?.instance?.value !== 'tcpSocket') || ($self.value !== '' && $self.value !== 0) {{`}}`}}"
        message: {{ i18n "值不能为零" .lang }}
  path:
    title: {{ i18n "请求路径" .lang }}
    type: string
    ui:rules:
      - maxLength250
      - validator: "{{`{{`}} !$widgetNode?.getSibling('enabled')?.instance?.value || $widgetNode?.getSibling('type')?.instance?.value !== 'httpGet' || $self.value !== '' {{`}}`}}"
        message: {{ i18n "值不能为空" .lang }}
  command:
    items:
      title: {{ i18n "命令" .lang }}
      type: string
      ui:rules:
        - required
        - maxLength128
    title: {{ i18n "命令" .lang }}
    type: array
    ui:component:
      name: bfArray
    ui:rules:
      - validator: "{{`{{`}} $widgetNode?.getSibling('type')?.instance?.value !== 'exec' || $self.value.length > 0 {{`}}`}}"
        message: {{ i18n "至少包含一条命令" .lang }}
  initialDelaySecs:
    title: {{ i18n "初始延时" .lang }}
    type: integer
    default: 0
    ui:component:
      name: bfInput
      props:
        max: 86400
        unit: s
  periodSecs:
    title: {{ i18n "检查间隔" .lang }}
    type: integer
    default: 10
    ui:component:
      name: bfInput
      props:
        max: 86400
        unit: s
  timeoutSecs:
    title: {{ i18n "超时时间" .lang }}
    type: integer
    default: 1
    ui:component:
      name: bfInput
      props:
        max: 86400
        unit: s
  successThreshold:
    title: {{ i18n "成功阈值" .lang }}
    type: integer
    default: 1
    ui:component:
      props:
        max: 2048
  failureThreshold:
    title: {{ i18n "失败阈值" .lang }}
    type: integer
    default: 3
    ui:component:
      props:
        max: 2048
{{- end }}

{{- define "container.resource" }}
resource:
  title: {{ i18n "资源" .lang }}
  type: object
  properties:
    requests:
      type: object
      properties:
        cpu:
          title: {{ i18n "CPU 预留" .lang }}
          type: integer
          ui:component:
            name: bfInput
            props:
              max: 256000
              unit: mCPUs
          ui:props:
            labelWidth: 200
        memory:
          title: {{ i18n "内存预留" .lang }}
          type: integer
          ui:component:
            name: bfInput
            props:
              max: 256000
              unit: MiB
        ephemeral-storage:
          title: {{ i18n "临时存储预留" .lang }}
          type: integer
          ui:component:
            name: bfInput
            props:
              max: 256000
              unit: GiB
        extra:
          title: {{ i18n "自定义资源预留" .lang }}
          type: array
          maxItems: 3
          items:
            type: object
            properties:
              key:
                title: {{ i18n "资源类型" .lang }}
                type: string
                ui:component:
                  name: select
                  props:
                    clearable: false
                    datasource:
                      - label: {{ i18n "共享网卡（eip）" .lang }}
                        value: tke.cloud.tencent.com/eip
                      - label: {{ i18n "独立网卡（eni-ip）" .lang }}
                        value: tke.cloud.tencent.com/eni-ip
                      - label: {{ i18n "算力-GPU" .lang }}
                        value: tencent.com/fgpu
                      - label: nvidia.com/gpu
                        value: nvidia.com/gpu
                      - label: huawei.com/Ascend910
                        value: huawei.com/Ascend910
                ui:rules:
                  - required
                  - validator: "{{`{{`}} $self.widgetNode.parent.parent.children.every(node => node.children[0].instance === $self || node.children[0].instance.value !== $self.value) {{`}}`}}"
                    message: {{ i18n "同类资源不可重复设置配额" .lang }}
              value:
                title: {{ i18n "值" .lang }}
                type: string
                ui:rules:
                  - required
                  - maxLength64
                  - numberRegex
          ui:component:
            name: bfArray
    limits:
      type: object
      properties:
        cpu:
          title: {{ i18n "CPU 限制" .lang }}
          type: integer
          ui:component:
            name: bfInput
            props:
              max: 256000
              unit: mCPUs
          ui:props:
            labelWidth: 200
        memory:
          title: {{ i18n "内存限制" .lang }}
          type: integer
          ui:component:
            name: bfInput
            props:
              max: 256000
              unit: MiB
        ephemeral-storage:
          title: {{ i18n "临时存储限制" .lang }}
          type: integer
          ui:component:
            name: bfInput
            props:
              max: 256000
              unit: GiB
        extra:
          title: {{ i18n "自定义资源限制" .lang }}
          type: array
          maxItems: 3
          items:
            type: object
            properties:
              key:
                title: {{ i18n "资源类型" .lang }}
                type: string
                ui:component:
                  name: select
                  props:
                    clearable: false
                    datasource:
                      - label: {{ i18n "共享网卡（eip）" .lang }}
                        value: tke.cloud.tencent.com/eip
                      - label: {{ i18n "独立网卡（eni-ip）" .lang }}
                        value: tke.cloud.tencent.com/eni-ip
                      - label: {{ i18n "算力-GPU" .lang }}
                        value: tencent.com/fgpu
                ui:rules:
                  - required
                  - validator: "{{`{{`}} $self.widgetNode.parent.parent.children.every(node => node.children[0].instance === $self || node.children[0].instance.value !== $self.value) {{`}}`}}"
                    message: {{ i18n "同类资源不可重复设置配额" .lang }}
              value:
                title: {{ i18n "值" .lang }}
                type: string
                ui:rules:
                  - required
                  - maxLength64
                  - numberRegex
          ui:component:
            name: bfArray
{{- end }}

{{- define "container.security" }}
security:
  title: {{ i18n "安全" .lang }}
  type: object
  properties:
    privileged:
      title: {{ i18n "特权模式" .lang }}
      type: boolean
    allowPrivilegeEscalation:
      title: {{ i18n "允许提权" .lang }}
      type: boolean
    runAsNonRoot:
      title: {{ i18n "以非 Root 运行" .lang }}
      type: boolean
    readOnlyRootFilesystem:
      title: {{ i18n "只读 Root 文件系统" .lang }}
      type: boolean
    runAsUser:
      title: {{ i18n "用户" .lang }}
      type: integer
      ui:component:
        props:
          max: 65535
    runAsGroup:
      title: {{ i18n "用户组" .lang }}
      type: integer
      ui:component:
        props:
          max: 65535
    procMount:
      title: {{ i18n "掩码挂载" .lang }}
      type: string
      ui:rules:
        - maxLength64
    capabilities:
      type: object
      properties:
        add:
          title: {{ i18n "新增权限" .lang }}
          type: array
          items:
            enum:
              - ALL
              - AUDIT_CONTROL
              - AUDIT_WRITE
              - BLOCK_SUSPEND
              - CHOWN
              - DAC_OVERRIDE
              - DAC_READ_SEARCH
              - FOWNER
              - FSETID
              - IPC_LOCK
              - IPC_OWNER
              - KILL
              - LEASE
              - LINUX_IMMUTABLE
              - MAC_ADMIN
              - MAC_OVERRIDE
              - MKNOD
              - NET_ADMIN
              - NET_BIND_SERVICE
              - NET_BROADCAST
              - NET_RAW
              - SETFCAP
              - SETGID
              - SETPCAP
              - SETUID
              - SYSLOGSYS_ADMIN
              - SYS_BOOT
              - SYS_CHROOT
              - SYS_MODULE
              - SYS_NICE
              - SYS_PACCT
              - SYS_PTRACE
              - SYS_RAWIO
              - SYS_RESOURCE
              - SYS_TIME
              - SYS_TTY_CONFIG
              - WAKE_ALARM
            type: string
          ui:component:
            name: select
            props:
              multiple: true
          uniqueItems: true
        drop:
          title: {{ i18n "消减权限" .lang }}
          type: array
          items:
            enum:
              - ALL
              - AUDIT_CONTROL
              - AUDIT_WRITE
              - BLOCK_SUSPEND
              - CHOWN
              - DAC_OVERRIDE
              - DAC_READ_SEARCH
              - FOWNER
              - FSETID
              - IPC_LOCK
              - IPC_OWNER
              - KILL
              - LEASE
              - LINUX_IMMUTABLE
              - MAC_ADMIN
              - MAC_OVERRIDE
              - MKNOD
              - NET_ADMIN
              - NET_BIND_SERVICE
              - NET_BROADCAST
              - NET_RAW
              - SETFCAP
              - SETGID
              - SETPCAP
              - SETUID
              - SYSLOGSYS_ADMIN
              - SYS_BOOT
              - SYS_CHROOT
              - SYS_MODULE
              - SYS_NICE
              - SYS_PACCT
              - SYS_PTRACE
              - SYS_RAWIO
              - SYS_RESOURCE
              - SYS_TIME
              - SYS_TTY_CONFIG
              - WAKE_ALARM
            type: string
          ui:component:
            name: select
            props:
              multiple: true
          uniqueItems: true
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

{{- define "container.mount" }}
mount:
  title: {{ i18n "挂载点" .lang }}
  type: object
  properties:
    volumes:
      title: {{ i18n "卷" .lang }}
      type: array
      items:
        type: object
        properties:
          name:
            title: {{ i18n "数据卷名称" .lang }}
            type: string
            ui:rules:
              - required
              - maxLength64
          mountPath:
            title: {{ i18n "挂载路径" .lang }}
            type: string
            ui:rules:
              - required
              - maxLength128
          subPath:
            title: {{ i18n "卷内子路径" .lang }}
            type: string
            ui:rules:
              - maxLength128
          readOnly:
            title: {{ i18n "只读" .lang }}
            type: boolean
            ui:component:
              name: checkbox
      ui:component:
        name: bfArray
      ui:props:
        showTitle: false
{{- end }}

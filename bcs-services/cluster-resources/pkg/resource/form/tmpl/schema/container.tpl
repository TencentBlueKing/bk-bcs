{{- define "container.containerGroup" }}
containerGroup:
  title: 容器组
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
{{- end }}

{{- define "container.initContainers" }}
initContainers:
  title: 初始容器
  type: array
  items:
    type: object
    properties:
      {{- include "container.basic" . | indent 6 }}
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
  ui:group:
    props:
      showTitle: false
{{- end }}

{{- define "container.containers" }}
containers:
  title: 标准容器
  type: array
  items:
    type: object
    properties:
      {{- include "container.basic" . | indent 6 }}
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
  ui:group:
    props:
      showTitle: false
{{- end }}

{{- define "container.basic" }}
basic:
  title: 基础信息
  type: object
  required:
    - name
    - image
  properties:
    image:
      title: 容器镜像
      type: string
    name:
      title: 容器名称
      type: string
    pullPolicy:
      title: 拉取策略
      type: string
      ui:component:
        name: select
        props:
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
  title: 命令
  type: object
  properties:
    workingDir:
      title: 工作目录
      type: string
    stdin:
      title: 标准输入
      type: boolean
    stdinOnce:
      title: 仅一次
      type: boolean
    tty:
      title: tty
      type: boolean
    command:
      title: 命令
      type: array
      items:
        type: string
      ui:component:
        name: noTitleArray
    args:
      title: 参数
      type: array
      items:
        type: string
      ui:component:
        name: noTitleArray
{{- end }}

{{- define "container.service" }}
service:
  title: 服务端口
  type: object
  properties:
    ports:
      type: array
      items:
        type: object
        properties:
          name:
            title: 名称
            type: string
          containerPort:
            title: 容器端口
            type: integer
            ui:component:
              props:
                max: 65535
          protocol:
            title: 协议
            type: string
            ui:component:
              name: select
              props:
                datasource:
                  - label: TCP
                    value: TCP
                  - label: UDP
                    value: UDP
          hostPort:
            title: 主机端口
            type: integer
            ui:component:
              props:
                max: 65535
      ui:component:
        name: noTitleArray
      ui:props:
        showTitle: false
{{- end }}

{{- define "container.envs" }}
envs:
  title: 环境变量
  type: object
  properties:
    vars:
      type: array
      items:
        type: object
        properties:
          type:
            title: 类型
            type: string
            ui:component:
              name: select
              props:
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
          name:
            title: 内容（Name/Prefix）
            type: string
          source:
            title: 来源
            type: string
          value:
            title: 值
            type: string
      ui:component:
        name: noTitleArray
      ui:props:
        showTitle: false
{{- end }}

{{- define "container.healthz" }}
healthz:
  title: 健康检查
  type: object
  properties:
    readinessProbe:
      title: 就绪探针
      type: object
      {{- include "container.probe" . | indent 6 }}
    livenessProbe:
      title: 存活探针
      type: object
      {{- include "container.probe" . | indent 6 }}
  ui:group:
    name: collapse
{{- end }}

{{- define "container.probe" }}
properties:
  type:
    title: 检查类型
    type: string
    ui:component:
      name: select
      props:
        datasource:
          - label: httpGet
            value: httpGet
          - label: tcpSocket
            value: tcpSocket
          - label: exec
            value: exec
  port:
    title: 端口
    type: integer
    ui:component:
      props:
        max: 65535
  path:
    title: 请求路径
    type: string
  initialDelaySecs:
    title: 初始延时
    type: integer
    ui:component:
      name: unitInput
      props:
        max: 86400
        unit: s
  periodSecs:
    title: 检查间隔
    type: integer
    ui:component:
      name: unitInput
      props:
        max: 86400
        unit: s
  timeoutSecs:
    title: 超时时间
    type: integer
    ui:component:
      name: unitInput
      props:
        max: 86400
        unit: s
  successThreshold:
    title: 成功阈值
    type: integer
  failureThreshold:
    title: 失败阈值
    type: integer
  command:
    items:
      title: 命令
      type: string
    title: 命令
    type: array
    ui:component:
      name: noTitleArray
{{- end }}

{{- define "container.resource" }}
resource:
  title: 资源
  type: object
  properties:
    requests:
      type: object
      properties:
        cpu:
          title: CPU 预留
          type: integer
          ui:component:
            name: unitInput
            props:
              unit: mCPUs
          ui:props:
            labelWidth: 200
        memory:
          title: 内存预留
          type: integer
          ui:component:
            name: unitInput
            props:
              unit: Mi
    limits:
      type: object
      properties:
        cpu:
          title: CPU 限制
          type: integer
          ui:component:
            name: unitInput
            props:
              unit: mCPUs
          ui:props:
            labelWidth: 200
        memory:
          title: 内存限制
          type: integer
          ui:component:
            name: unitInput
            props:
              unit: Mi
{{- end }}

{{- define "container.security" }}
security:
  properties:
    title: 安全
    type: object
    privileged:
      title: 特权模式
      type: boolean
    allowPrivilegeEscalation:
      title: 允许提权
      type: boolean
    runAsNonRoot:
      title: 以非 Root 运行
      type: boolean
    readOnlyRootFilesystem:
      title: 只读 Root 文件系统
      type: boolean
    runAsUser:
      title: 用户
      type: integer
    runAsGroup:
      title: 用户组
      type: integer
    procMount:
      title: 掩码挂载
      type: string
    capabilities:
      type: object
      properties:
        add:
          title: 新增权限
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
          title: 消减权限
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

{{- define "container.mount" }}
mount:
  title: 挂载点
  type: object
  properties:
    volumes:
      title: 卷
      type: array
      items:
        type: object
        properties:
          name:
            title: 数据卷名称
            type: string
          mountPath:
            title: 挂载路径
            type: string
          subPath:
            title: 卷内子路径
            type: string
          readOnly:
            title: 只读
            type: boolean
            ui:component:
              name: checkbox
      ui:component:
        name: noTitleArray
      ui:props:
        showTitle: false
{{- end }}

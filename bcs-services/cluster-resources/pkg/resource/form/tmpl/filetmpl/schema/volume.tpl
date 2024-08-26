{{- define "volume.pvc" }}
pvc:
  title: PVC
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
          - nameRegexWithVar
      pvcName:
        title: PersistentVolumeClaim
        type: string
        ui:component:
          props:
            clearable: false
        ui:reactions:
          - lifetime: init
            then:
              actions:
                - "{{`{{`}} $loadDataSource {{`}}`}}"
          - source: "metadata.namespace"
            then:
              actions:
                - "{{`{{`}} $loadDataSource {{`}}`}}"
        ui:rules:
          - required
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

{{- define "volume.hostPath" }}
hostPath:
  title: HostPath
  type: array
  items:
    type: object
    properties:
      name:
        title: {{ i18n "名称" .lang }}
        type: string
        ui:rules:
          - required
          - maxLength128
          - nameRegexWithVar
      path:
        title: {{ i18n "路径或节点" .lang }}
        type: string
        ui:rules:
          - required
          - maxLength250
      type:
        title: {{ i18n "类型" .lang }}
        type: string
        default: Directory
        ui:component:
          name: select
          props:
            clearable: false
            datasource:
              - label: DirectoryOrCreate
                value: DirectoryOrCreate
              - label: Directory
                value: Directory
              - label: FileOrCreate
                value: FileOrCreate
              - label: File
                value: File
              - label: Socket
                value: Socket
              - label: CharDevice
                value: CharDevice
              - label: BlockDevice
                value: BlockDevice
  ui:component:
    name: bfArray
  ui:props:
    showTitle: false
{{- end }}

{{- define "volume.configMap" }}
configMap:
  title: ConfigMap
  type: array
  items:
    type: object
    required:
      - name
      - cmName
    properties:
      name:
        title: {{ i18n "名称" .lang }}
        type: string
        ui:rules:
          - required
          - maxLength128
          - nameRegexWithVar
      defaultMode:
        title: {{ i18n "默认模式" .lang }}
        type: string
        default: "0644"
        description: {{ i18n "八进制数字（0000-0777）或十进制数字（0-511）" .lang }}
        ui:rules:
          - numberRegex
      cmName:
        title: ConfigMap
        type: string
        ui:component:
          props:
            clearable: false
        ui:rules:
          - required
      items:
        title: Items
        type: array
        items:
          properties:
            key:
              title: {{ i18n "键" .lang }}
              type: string
              ui:rules:
                - required
                - maxLength128
            path:
              title: {{ i18n "映射目标路径" .lang }}
              type: string
              ui:rules:
                - required
                - maxLength128
          type: object
        ui:component:
          name: bfArray
    ui:group:
      props:
        showTitle: false
        type: normal
      style:
        background: '#fff'
  ui:group:
    props:
      showTitle: false
{{- end }}

{{- define "volume.secret" }}
secret:
  title: Secret
  type: array
  items:
    type: object
    required:
      - name
      - secretName
    properties:
      name:
        title: {{ i18n "名称" .lang }}
        type: string
        ui:rules:
          - required
          - maxLength128
          - nameRegexWithVar
      defaultMode:
        title: {{ i18n "默认模式" .lang }}
        type: string
        default: "0644"
        description: {{ i18n "八进制数字（0000-0777）或十进制数字（0-511）" .lang }}
        ui:rules:
          - numberRegex
      secretName:
        title: Secret
        type: string
        ui:component:
          props:
            clearable: false
        ui:rules:
          - required
      items:
        title: Items
        type: array
        items:
          properties:
            key:
              title: {{ i18n "键" .lang }}
              type: string
              ui:rules:
                - required
                - maxLength128
            path:
              title: {{ i18n "映射目标路径" .lang }}
              type: string
              ui:rules:
                - required
                - maxLength128
          type: object
        ui:component:
          name: bfArray
    ui:group:
      props:
        showTitle: false
        type: normal
      style:
        background: '#fff'
  ui:group:
    props:
      showTitle: false
{{- end }}

{{- define "volume.emptyDir" }}
emptyDir:
  title: EmptyDir
  type: array
  items:
    type: object
    properties:
      name:
        title: {{ i18n "名称" .lang }}
        type: string
        ui:rules:
          - required
          - maxLength128
          - nameRegexWithVar
  ui:component:
    name: bfArray
  ui:props:
    showTitle: false
{{- end }}

{{- define "volume.nfs" }}
nfs:
  title: NFS
  type: array
  items:
    type: object
    properties:
      name:
        title: {{ i18n "名称" .lang }}
        type: string
        ui:rules:
          - required
          - maxLength128
          - nameRegexWithVar
      path:
        title: {{ i18n "路径" .lang }}
        type: string
        ui:rules:
          - required
          - maxLength128
      server:
        title: Server
        type: string
        ui:rules:
          - required
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

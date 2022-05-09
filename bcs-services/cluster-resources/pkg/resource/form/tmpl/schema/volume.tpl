{{- define "volume.pvc" }}
pvc:
  title: PVC
  type: array
  items:
    type: object
    properties:
      name:
        title: 名称
        type: string
      pvcName:
        title: PersistentVolumeClaim
        type: string
        ui:component:
          name: select
          props:
            clearable: false
            searchable: true
            remoteConfig:
              params:
                format: selectItems
              url: "{{`{{`}} `${$context.baseUrl}/projects/${$context.projectID}/clusters/${$context.clusterID}/namespaces/${$self.getValue('metadata.namespace')}/storages/persistent_volume_claims` {{`}}`}}"
        ui:reactions:
          - lifetime: init
            then:
              actions:
                - "{{`{{`}} $loadDataSource {{`}}`}}"
          - source: "metadata.namespace"
            then:
              actions:
                - "{{`{{`}} $loadDataSource {{`}}`}}"
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

{{- define "volume.hostPath" }}
hostPath:
  title: HostPath
  type: array
  items:
    type: object
    properties:
      name:
        title: 名称
        type: string
      path:
        title: 路径或节点
        type: string
      type:
        title: 类型
        type: string
        ui:component:
          name: select
          props:
            clearable: true
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
    name: noTitleArray
  ui:props:
    showTitle: false
{{- end }}

{{- define "volume.configMap" }}
configMap:
  title: ConfigMap
  type: array
  items:
    type: object
    properties:
      name:
        title: 名称
        type: string
      defaultMode:
        description: 三位数字，如 644
        title: 默认模式
        type: integer
      cmName:
        title: ConfigMap
        type: string
        ui:component:
          name: select
          props:
            clearable: false
            searchable: true
            remoteConfig:
              params:
                format: selectItems
              url: "{{`{{`}} `${$context.baseUrl}/projects/${$context.projectID}/clusters/${$context.clusterID}/namespaces/${$self.getValue('metadata.namespace')}/configs/configmaps` {{`}}`}}"
        ui:reactions:
          - lifetime: init
            then:
              actions:
                - "{{`{{`}} $loadDataSource {{`}}`}}"
          - source: "metadata.namespace"
            then:
              actions:
                - "{{`{{`}} $loadDataSource {{`}}`}}"
      items:
        title: Items
        type: array
        items:
          properties:
            key:
              title: 键
              type: string
            path:
              title: 映射目标路径
              type: string
          type: object
        ui:component:
          name: noTitleArray
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
    properties:
      name:
        title: 名称
        type: string
      defaultMode:
        title: 默认模式
        type: integer
        description: 三位数字，如 644
      secretName:
        title: Secret
        type: string
        ui:component:
          name: select
          props:
            clearable: false
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
      items:
        title: Items
        type: array
        items:
          properties:
            key:
              title: 键
              type: string
            path:
              title: 映射目标路径
              type: string
          type: object
        ui:component:
          name: noTitleArray
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
        title: 名称
        type: string
  ui:component:
    name: noTitleArray
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
        title: 名称
        type: string
      path:
        title: 路径
        type: string
      server:
        title: Server
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

{{- define "config.cmData" }}
data:
  title: {{ i18n "数据" .lang }}
  type: object
  properties:
    immutable:
      title: {{ i18n "不可变更" .lang }}
      type: boolean
      description: {{ i18n "（k8s 1.19+） 保护应用，使之免受意外更新所带来的负面影响;<br>降低对 kube-apiserver 的性能压力，系统会关闭对已标记为不可变更的资源的监视操作" .lang }}
      ui:component:
        props:
          visible: {{ .featureGates.ImmutableEphemeralVolumes }}
    items:
      type: array
      items:
        type: object
        required:
          - key
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
            ui:component:
              name: bfInput
              props:
                placeholder: {{ i18n "值（支持多行文本）" .lang }}
                maxRows: 6
        ui:group:
          props:
            showTitle: false
            type: normal
      ui:component:
        name: bfArray
      ui:props:
        showTitle: false
  ui:group:
    props:
      border: true
      showTitle: true
      type: card
{{- end }}

{{- define "config.secretData" }}
data:
  title: {{ i18n "数据" .lang }}
  type: object
  properties:
    type:
      title: {{ i18n "类型" .lang }}
      type: string
      default: Opaque
      ui:component:
        name: select
        props:
          clearable: false
          # 更新时无法修改 Secret 类型
          disabled: {{ eq .action "update" }}
          datasource:
            - label: Opaque
              value: Opaque
            - label: Docker Registry
              value: kubernetes.io/dockerconfigjson
            - label: BasicAuth
              value: kubernetes.io/basic-auth
            - label: SSHAuth
              value: kubernetes.io/ssh-auth
            - label: TLS
              value: kubernetes.io/tls
            - label: ServiceAccount Token
              value: kubernetes.io/service-account-token
      ui:reactions:
        - target: data.opaque
          if: "{{`{{`}} $self.value === 'Opaque' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              value: []
              visible: false
        - target: data.docker.registry
          if: "{{`{{`}} $self.value === 'kubernetes.io/dockerconfigjson' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
        - target: data.docker.username
          if: "{{`{{`}} $self.value === 'kubernetes.io/dockerconfigjson' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
        - target: data.docker.password
          if: "{{`{{`}} $self.value === 'kubernetes.io/dockerconfigjson' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
        - target: data.basicAuth.username
          if: "{{`{{`}} $self.value === 'kubernetes.io/basic-auth' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
        - target: data.basicAuth.password
          if: "{{`{{`}} $self.value === 'kubernetes.io/basic-auth' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
        - target: data.sshAuth.publicKey
          if: "{{`{{`}} $self.value === 'kubernetes.io/ssh-auth' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
        - target: data.sshAuth.privateKey
          if: "{{`{{`}} $self.value === 'kubernetes.io/ssh-auth' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
        - target: data.tls.cert
          if: "{{`{{`}} $self.value === 'kubernetes.io/tls' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
        - target: data.tls.privateKey
          if: "{{`{{`}} $self.value === 'kubernetes.io/tls' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
        - target: data.saToken.namespace
          if: "{{`{{`}} $self.value === 'kubernetes.io/service-account-token' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
        - target: data.saToken.saName
          if: "{{`{{`}} $self.value === 'kubernetes.io/service-account-token' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
        - target: data.saToken.token
          if: "{{`{{`}} $self.value === 'kubernetes.io/service-account-token' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
        - target: data.saToken.cert
          if: "{{`{{`}} $self.value === 'kubernetes.io/service-account-token' {{`}}`}}"
          then:
            state:
              visible: true
          else:
            state:
              visible: false
    immutable:
      title: {{ i18n "不可变更" .lang }}
      type: boolean
      description: {{ i18n "（k8s 1.19+） 保护应用，使之免受意外更新所带来的负面影响;<br>降低对 kube-apiserver 的性能压力，系统会关闭对已标记为不可变更的资源的监视操作" .lang }}
      ui:component:
        props:
          visible: {{ .featureGates.ImmutableEphemeralVolumes }}
    opaque:
      title: Opaque
      type: array
      items:
        type: object
        required:
          - key
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
            ui:component:
              name: bfInput
              props:
                placeholder: {{ i18n "值（明文即可，提交后会自动转换为 base64 编码）" .lang }}
                maxRows: 6
        ui:group:
          props:
            showTitle: false
            type: normal
      ui:component:
        name: bfArray
    docker:
      type: object
      required:
        - registry
        - username
      properties:
        registry:
          title: {{ i18n "仓库地址" .lang }}
          type: string
          ui:rules:
            - validator: "{{`{{`}} $self.getValue('data.type') !== 'kubernetes.io/dockerconfigjson' || $self.value !== '' {{`}}`}}"
              message: {{ i18n "值不能为空" .lang }}
            - maxLength128
        username:
          title: {{ i18n "用户名" .lang }}
          type: string
          ui:rules:
            - validator: "{{`{{`}} $self.getValue('data.type') !== 'kubernetes.io/dockerconfigjson' || $self.value !== '' {{`}}`}}"
              message: {{ i18n "值不能为空" .lang }}
            - maxLength64
        password:
          title: {{ i18n "密码" .lang }}
          type: string
          ui:component:
            props:
              type: password
          ui:rules:
            - maxLength128
    basicAuth:
      type: object
      required:
        - username
      properties:
        username:
          title: {{ i18n "用户名" .lang }}
          type: string
          ui:rules:
            - validator: "{{`{{`}} $self.getValue('data.type') !== 'kubernetes.io/basic-auth' || $self.value !== '' {{`}}`}}"
              message: {{ i18n "值不能为空" .lang }}
            - maxLength64
        password:
          title: {{ i18n "密码" .lang }}
          type: string
          ui:component:
            props:
              type: password
          ui:rules:
            - maxLength128
    sshAuth:
      type: object
      required:
        - publicKey
        - privateKey
      properties:
        publicKey:
          title: {{ i18n "公钥" .lang }}
          type: string
          ui:component:
            name: bfInput
            props:
              placeholder: {{ i18n "值（明文即可，提交后会自动转换为 base64 编码）" .lang }}
              type: textarea
              rows: 10
          ui:rules:
            - validator: "{{`{{`}} $self.getValue('data.type') !== 'kubernetes.io/ssh-auth' || $self.value !== '' {{`}}`}}"
              message: {{ i18n "值不能为空" .lang }}
        privateKey:
          title: {{ i18n "私钥" .lang }}
          type: string
          ui:component:
            name: bfInput
            props:
              placeholder: {{ i18n "值（明文即可，提交后会自动转换为 base64 编码）" .lang }}
              type: textarea
              rows: 10
          ui:rules:
            - validator: "{{`{{`}} $self.getValue('data.type') !== 'kubernetes.io/ssh-auth' || $self.value !== '' {{`}}`}}"
              message: {{ i18n "值不能为空" .lang }}
    tls:
      type: object
      required:
        - cert
      properties:
        cert:
          title: {{ i18n "证书" .lang }}
          type: string
          ui:component:
            name: bfInput
            props:
              placeholder: {{ i18n "值（明文即可，提交后会自动转换为 base64 编码）" .lang }}
              type: textarea
              rows: 10
          ui:rules:
            - validator: "{{`{{`}} $self.getValue('data.type') !== 'kubernetes.io/tls' || $self.value !== '' {{`}}`}}"
              message: {{ i18n "值不能为空" .lang }}
        privateKey:
          title: {{ i18n "私钥" .lang }}
          type: string
          ui:component:
            name: bfInput
            props:
              placeholder: {{ i18n "值（明文即可，提交后会自动转换为 base64 编码）" .lang }}
              type: textarea
              rows: 10
    saToken:
      type: object
      required:
        - namespace
        - saName
        - token
      properties:
        namespace:
          title: {{ i18n "命名空间" .lang }}
          type: string
          default: {{ .namespace }}
          ui:component:
            name: select
            props:
              clearable: false
          ui:rules:
            - validator: "{{`{{`}} $self.getValue('data.type') !== 'kubernetes.io/service-account-token' || $self.value !== '' {{`}}`}}"
              message: {{ i18n "值不能为空" .lang }}
        saName:
          title: {{ i18n "服务账号" .lang }}
          type: string
          ui:component:
            name: select
            props:
              clearable: true
          ui:rules:
            - validator: "{{`{{`}} $self.getValue('data.type') !== 'kubernetes.io/service-account-token' || $self.value !== '' {{`}}`}}"
              message: {{ i18n "值不能为空" .lang }}
        token:
          title: Token
          type: string
          ui:component:
            props:
              placeholder: {{ i18n "值（明文即可，提交后会自动转换为 base64 编码）" .lang }}
          ui:rules:
            - validator: "{{`{{`}} $self.getValue('data.type') !== 'kubernetes.io/service-account-token' || $self.value !== '' {{`}}`}}"
              message: {{ i18n "值不能为空" .lang }}
        cert:
          title: {{ i18n "证书" .lang }}
          type: string
          description: {{ i18n "证书由 kubernetes 生成，不允许编辑" .lang }}
          ui:component:
            name: bfInput
            props:
              # 证书由 k8s 生成，不允许编辑
              placeholder: {{ i18n "证书由 kubernetes 生成，不允许编辑" .lang }}
              disabled: true
              type: textarea
              rows: 10
  ui:group:
    props:
      border: true
      showTitle: true
      type: card
      verifiable: true
      hideEmptyRow: true
{{- end }}

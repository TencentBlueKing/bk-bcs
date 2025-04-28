{{- define "bscpconfig.Data" }}
spec:
  title: {{ i18n "配置信息" .lang }}
  type: object
  properties:
    provider:
      title: {{ i18n "provider" .lang }}
      type: object
      required:
        - feedAddr
        - biz
        - token      
        - app      
      properties:
        feedAddr: 
          title: {{ i18n "feedAddr" .lang }}
          type: string
          ui:component:
            name: bfInput
            props:
              placeholder: {{ i18n "feed 地址" .lang }}
              maxRows: 6
          ui:rules:
            - required
        biz: 
          title: {{ i18n "biz" .lang }}
          type: integer
          ui:component:
            name: bfInput
            props:
              placeholder: {{ i18n "业务ID" .lang }}
              maxRows: 6
          ui:rules:
            - required
        token: 
          title: {{ i18n "token" .lang }}
          type: string
          ui:component:
            name: bfInput
            props:
              placeholder: {{ i18n "客户端密钥" .lang }}
              maxRows: 6
          ui:rules:
            - required
        app: 
          title: {{ i18n "app" .lang }}
          type: string
          ui:component:
            name: bfInput
            props:
              placeholder: {{ i18n "服务名称" .lang }}
              maxRows: 6
          ui:rules:
            - required
    configSyncer:
      title: {{ i18n "configSyncer" .lang }}
      type: array
      items:
        type: object
        required:
          - configmapName
          - secretName      
          - secretType         
          - matchConfigs         
          - configData         
        properties:
          resourceType:
            title: {{ i18n "资源类型" .lang }}
            type: string
            default: configmap
            ui:component:
              name: radio
              props:
                datasource:
                  - label: {{ i18n "configmap" .lang }}
                    value: configmap
                  - label: {{ i18n "secret" .lang }}
                    value: secret
            ui:reactions:
              - target: "{{`{{`}} $widgetNode?.getSibling('configmapName')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'configmap' {{`}}`}}"
                then:
                  state:
                    visible: true
                else:
                  state:
                    visible: false
                    disable: true
              - target: "{{`{{`}} $widgetNode?.getSibling('secretName')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'secret' {{`}}`}}"
                then:
                  state:
                    visible: true
                else:
                  state:
                    visible: false
                    disable: true
              - target: "{{`{{`}} $widgetNode?.getSibling('secretType')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'secret' {{`}}`}}"
                then:
                  state:
                    visible: true
                else:
                  state:
                    visible: false
          configmapName:
            title: {{ i18n "name" .lang }}
            type: string 
            ui:component:
              name: bfInput
              props:
                placeholder: {{ i18n "生成configmap资源的名称" .lang }}
                maxRows: 6
            ui:rules:
              - validator: "{{`{{`}} $widgetNode?.getSibling('resourceType')?.instance?.value === 'secret' || $self.value !== '' {{`}}`}}"
                message: {{ i18n "值不能为空" .lang }}
          secretName:
            title: {{ i18n "name" .lang }}
            type: string
            ui:component:
              name: bfInput
              props:
                placeholder: {{ i18n "生成secret资源的名称" .lang }}
                maxRows: 6
            ui:rules:
              - validator: "{{`{{`}} $widgetNode?.getSibling('resourceType')?.instance?.value === 'configmap' || $self.value !== '' {{`}}`}}"
                message: {{ i18n "值不能为空" .lang }}
          secretType:
            title: {{ i18n "type" .lang }}
            type: string  
            default: Opaque
            ui:component:
              name: select
              props:
                clearable: false
                datasource:
                  - label: Opaque
                    value: Opaque
                  - label: kubernetes.io/service-account-token
                    value: kubernetes.io/service-account-token
                  - label: kubernetes.io/dockerconfigjson
                    value: kubernetes.io/dockerconfigjson     
                  - label: kubernetes.io/basic-auth
                    value: kubernetes.io/basic-auth    
                  - label: kubernetes.io/ssh-auth
                    value: kubernetes.io/ssh-auth   
                  - label: kubernetes.io/tls
                    value: kubernetes.io/tls   
                  - label: bootstrap.kubernetes.io/token
                    value: bootstrap.kubernetes.io/token                 
          associationRules:
            title: {{ i18n "关联规则" .lang }}
            type: string
            default: matchConfigs
            ui:component:
              name: radio
              props:
                datasource:
                  - label: {{ i18n "多个模糊匹配" .lang }}
                    value: matchConfigs
                  - label: {{ i18n "精准匹配" .lang }}
                    value: data
            ui:reactions:
              - target: "{{`{{`}} $widgetNode?.getSibling('matchConfigs')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'matchConfigs' {{`}}`}}"
                then:
                  state:
                    visible: true
                else:
                  state:
                    visible: false
                    value: []
                    disable: true                                                                                                
              - target: "{{`{{`}} $widgetNode?.getSibling('configData')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'data' {{`}}`}}"
                then:
                  state:
                    visible: true
                else:
                  state:
                    visible: false
                    value: []
                    disable: true
          matchConfigs:
            title: {{ i18n "matchConfigs" .lang }}
            type: array
            items:
              type: object
              required:
                - value
              properties:
                value:
                  title: {{ i18n "value" .lang }}
                  type: string
                  ui:component:
                    name: bfInput
                    props:
                      placeholder: {{ i18n "关联的bscp配置匹配规则，支持linux wilecard语法" .lang }}
                      maxRows: 6
                  ui:rules:
                    - required 
                  ui:props:
                    showTitle: false         
            ui:component:
              name: bfArray                                                                                                                                                                                    
          configData:
            title: {{ i18n "data" .lang }}
            type: array
            items:
              type: object
              required:
                - key
                - refConfig                  
              properties:
                key:
                  title: {{ i18n "key" .lang }}
                  type: string
                  ui:component:
                    name: bfInput
                    props:
                      placeholder: {{ i18n "名称" .lang }}
                      maxRows: 6
                  ui:rules:
                    - required
                  ui:props:
                    showTitle: false       
                refConfig:
                  title: {{ i18n "refConfig" .lang }}
                  type: string
                  ui:component:
                    name: bfInput
                    props:
                      placeholder: {{ i18n "关联的bscp配置项名称" .lang }}
                      maxRows: 6
                  ui:rules:
                    - required
                  ui:props:
                    showTitle: false                       
            ui:component:
              name: bfArray                                    
            ui:props:
              showTitle: false
              showTableHead: true                                              
        ui:order:
          - resourceType
          - configmapName
          - associationRules
          - secretName
          - secretType        
          - matchConfigs        
          - configData
        ui:group:
          props:
            border: true
            showTitle: false
            type: normal
          style:
            background: '#fff'
      ui:group:
        props:
          showTitle: false
  ui:group:
    name: collapse
    props:
      border: true
      showTitle: true
      type: card
      verifiable: true
      hideEmptyRow: true
      defaultActiveName:
        - provider          
        - configSyncer          
  ui:order:
    - provider
    - configSyncer            
{{- end }}

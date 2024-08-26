{{- define "workload.affinity" -}}
- - group:
      - - group:
            - [ "priority", "weight" ]
            - [ "selector" ]
          prop: nodeAffinity
      - - group:
            - [ "type", "priority", "namespaces", "weight" ]
            - [ "topologyKey" ]
            - - group:
                  - [ "expressions" ]
                  - [ "labels" ]
                prop: selector
          prop: podAffinity
          container:
            grid-template-columns: "1fr 1fr 1fr 100px"
    prop: affinity
{{- end }}

{{- define "workload.labels" }}
- - group:
      - [ "labels" ]
      - [ "templateLabels" ]
      - [ "jobTemplatelabels" ]
    prop: labels
{{- end }}

# 为 workload 污点/容忍添加 layout，限制运算符，影响，容忍时间的宽度
{{- define "workload.toleration" }}
- - group:
      - - group:
            - [ "key", "op", "value", "effect", "tolerationSecs", "." ]
          prop: rules
          container:
            grid-template-columns: "1fr 100px 1fr 245px 130px auto"
    prop: toleration
{{- end }}

{{- define "workload.networking" }}
- - group:
      - [ "dnsPolicy" ]
      - [ "hostIPC", "hostNetwork", "hostPID", "shareProcessNamespace" ]
      - [ "hostname", "subdomain" ]
      - [ "hostAliases", "." ]
      - [ "nameServers", "searches", "dnsResolverOpts", "dnsResolverOpts" ]
    prop: networking
{{- end }}

{{- define "workload.security" }}
- - group:
      - [ "runAsUser", "runAsNonRoot" ]
      - [ "runAsGroup", "fsGroup" ]
      - - group:
            - [ "level", "role" ]
            - [ "type", "user" ]
          prop: seLinuxOpt
    prop: security
{{- end }}

{{- define "workload.other" }}
- - group:
      - [ "restartPolicy", "terminationGracePeriodSecs" ]
      - [ "imagePullSecrets", "saName" ]
    prop: other
{{- end }}

{{- define "workload.readinessGates" }}
- - group:
      - [ "readinessGates" ]
    prop: readinessGates
    container:
      grid-template-columns: "1fr auto"
{{- end }}

{{- define "workload.volume" }}
- - group:
      - - group:
            - [ "name", "pvcName", "readOnly", "." ]
          prop: pvc
          container:
            grid-template-columns: "1fr 1fr 60px auto"
      - - group:
            - [ "name", "path", "type", "." ]
          prop: hostPath
          container:
            grid-template-columns: "1fr 1fr 150px auto"
      - - group:
            - [ "name", "defaultMode", "cmName" ]
            - [ "items" ]
          prop: configMap
      - - group:
            - [ "name", "defaultMode", "secretName" ]
            - [ "items" ]
          prop: secret
      - - group:
            - [ "name" ]
          prop: emptyDir
          container:
            grid-template-columns: "1fr auto"
      - - group:
            - [ "name", "path", "server", "readOnly" ]
          prop: nfs
          container:
            grid-template-columns: "1fr 1fr 1fr 60px auto"
    prop: volume
{{- end }}

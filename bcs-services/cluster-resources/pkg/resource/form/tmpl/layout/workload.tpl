{{- define "workload.affinity" -}}
- - group:
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
      - - group:
            - [ "priority", "weight" ]
            - [ "selector" ]
          prop: nodeAffinity
    prop: affinity
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

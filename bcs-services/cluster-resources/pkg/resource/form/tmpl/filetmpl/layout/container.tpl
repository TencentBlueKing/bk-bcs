{{- define "container.containerGroup" }}
- - group:
      {{- include "container.initContainers" . | indent 6 }}
      {{- include "container.containers" . | indent 6 }}
    prop: containerGroup
{{- end }}

{{- define "container.initContainers" }}
- - group:
      {{- include "container.basic" . | indent 6 }}
      {{- include "container.command" . | indent 6 }}
      {{- include "container.service" . | indent 6 }}
      {{- include "container.envs" . | indent 6 }}
      {{- include "container.resource" . | indent 6 }}
      {{- include "container.security" . | indent 6 }}
      {{- include "container.mount" . | indent 6 }}
    prop: initContainers
{{- end }}

{{- define "container.containers" }}
- - group:
      {{- include "container.basic" . | indent 6 }}
      {{- include "container.command" . | indent 6 }}
      {{- include "container.service" . | indent 6 }}
      {{- include "container.envs" . | indent 6 }}
      {{- include "container.healthz" . | indent 6 }}
      {{- include "container.resource" . | indent 6 }}
      {{- include "container.security" . | indent 6 }}
      {{- include "container.mount" . | indent 6 }}
    prop: containers
{{- end }}

{{- define "container.basic" }}
- - group:
      - [ "name" ]
      - [ "image", "pullPolicy" ]
    prop: basic
{{- end }}

{{- define "container.command" }}
- - group:
      - [ "workingDir", "stdin", "stdinOnce", "tty" ]
      - [ "command", "args" ]
    prop: command
{{- end }}

{{- define "container.service" }}
- - group:
      - - group:
            - [ "name", "containerPort", "protocol", "hostPort", "." ]
          prop: ports
          container:
            grid-template-columns: "1fr 1fr 1fr 1fr auto"
    prop: service
{{- end }}

{{- define "container.envs" }}
- - group:
      - - group:
            - [ "type", "name", "source", "value", "." ]
          prop: vars
          container:
            grid-template-columns: "1fr 1fr 1fr 1fr auto"
    prop: envs
{{- end }}

{{- define "container.healthz" }}
- - group:
      - - group:
            # 请求路径（path）占两列宽度
            - [ "enabled", "type", "port", "path", "path" ]
            - [ "command" ]
            - [ "initialDelaySecs", "periodSecs", "timeoutSecs", "successThreshold", "failureThreshold" ]
          prop: readinessProbe
      - - group:
            - [ "enabled", "type", "port", "path", "path" ]
            - [ "command" ]
            - [ "initialDelaySecs", "periodSecs", "timeoutSecs", "successThreshold", "failureThreshold" ]
          prop: livenessProbe
    prop: healthz
{{- end }}

{{- define "container.resource" }}
- - group:
      - - group:
            - [ "cpu" ]
            - [ "memory" ]
            - [ "ephemeral-storage" ]
            - - group:
                  - [ "key", "value", "." ]
                prop: extra
                container:
                  grid-template-columns: "1fr 1fr auto"
          prop: requests
        # NOTE 这里只有一个 - 表示 requests, limits 是同组的，同行布局
        - group:
            - [ "cpu" ]
            - [ "memory" ]
            - [ "ephemeral-storage" ]
            - - group:
                  - [ "key", "value", "." ]
                prop: extra
                container:
                  grid-template-columns: "1fr 1fr auto"
          prop: limits
    prop: resource
{{- end }}

{{- define "container.security" }}
- - group:
      - [ "allowPrivilegeEscalation", "privileged", "runAsNonRoot", "readOnlyRootFilesystem" ]
      - [ "runAsUser", "." ]
      - [ "runAsGroup", "procMount" ]
      - - group:
            - [ "add", "drop" ]
          prop: capabilities
      - - group:
            - [ "level", "role" ]
            - [ "type", "user" ]
          prop: seLinuxOpt
    prop: security
{{- end }}

{{- define "container.mount" }}
- - group:
      - - group:
            - [ "name", "mountPath", "subPath", "readOnly", "." ]
          prop: volumes
          container:
            grid-template-columns: "1fr 1fr 1fr 60px auto"
    prop: mount
{{- end }}

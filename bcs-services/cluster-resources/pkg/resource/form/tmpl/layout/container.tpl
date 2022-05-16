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
            # 请求路径（path）占三列宽度
            - [ "type", "port", "path", "path", "path" ]
            - [ "initialDelaySecs", "periodSecs", "timeoutSecs", "successThreshold", "failureThreshold" ]
            - [ "command" ]
          prop: readinessProbe
      - - group:
            - [ "type", "port", "path", "path", "path" ]
            - [ "initialDelaySecs", "periodSecs", "timeoutSecs", "successThreshold", "failureThreshold" ]
            - [ "command" ]
          prop: livenessProbe
    prop: healthz
{{- end }}

{{- define "container.resource" }}
- - group:
      - - group:
            - [ "cpu", "memory" ]
          prop: requests
      - - group:
            - [ "cpu", "memory" ]
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

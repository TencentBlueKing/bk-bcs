# -*- coding: utf-8 -*-
"""
Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community
Edition) available.
Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://opensource.org/licenses/MIT

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.
"""
import logging

from django.conf import settings

logger = logging.getLogger(__name__)

# K8S lb default name
K8S_LB_CHART_NAME = "blueking-nginx-ingress"
CONTROLLER_IMAGE_PATH = "public/bcs/k8s/nginx-ingress-controller"
BACKEND_IMAGE_PATH = "public/bcs/k8s/defaultbackend"

# k8s lb label
K8S_LB_LABEL = {"nodetype": "lb"}

# k8s nginx ingress controller helm chart values
K8S_NGINX_INGRESS_CONTROLLER_CHART_VALUES = """
controller:
  kind: Deployment
  name: controller
  image:
    repository: __REPO_ADDR__/__CONTROLLER_IMAGE_PATH__
    tag: "__TAG__"
    pullPolicy: IfNotPresent
  hostNetwork: true
  dnsPolicy: ClusterFirstWithHostNet
  scope:
    enabled: false
    namespace: __NAMESPACE__
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate
  minReadySeconds: 0
  tolerations:
  - key: "special"
    operator: "Equal"
    value: "true"
    effect: "NoSchedule"
  affinity:
    podAntiAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        - labelSelector:
            matchExpressions:
            - key: "app"
              operator: In
              values:
              - blueking-nginx-ingress-controller
          topologyKey: "kubernetes.io/hostname"
  nodeSelector:
    nodetype: lb
  livenessProbe:
    failureThreshold: 10
    initialDelaySeconds: 10
    periodSeconds: 10
    successThreshold: 1
    timeoutSeconds: 5
    port: 10254
  readinessProbe:
    failureThreshold: 10
    initialDelaySeconds: 10
    periodSeconds: 10
    successThreshold: 1
    timeoutSeconds: 5
    port: 10254
  podAnnotations:
    prometheus.io/port: '10254'
    prometheus.io/scrape: 'true'
  replicaCount: __CONTROLLER_REPLICA_COUNT__
  minAvailable: 1
  resources: {}
  ports:
    httpEnabled: __HTTP_ENABLED__
    httpPort: __HTTP_PORT__
    httpsEnabled: __HTTPS_ENABLED__
    httpsPort: __HTTPS_PORT__
defaultBackend:
  enabled: true
  name: default-backend
  image:
    repository: __REPO_ADDR__/__BACKEND_IMAGE_PATH__
    tag: "1.5"
    pullPolicy: IfNotPresent
  port: 8080
  tolerations:
  - key: "special"
    operator: "Equal"
    value: "true"
    effect: "NoSchedule"
  nodeSelector:
    nodetype: lb
  replicaCount: 1
  minAvailable: 1
  resources:
    limits:
      cpu: 10m
      memory: 20Mi
    requests:
      cpu: 10m
      memory: 20Mi
  service:
    servicePort: 80
    targetPort: 8080
rbac:
  create: true
serviceAccount:
  create: true
  name: serviceaccount
tcp:
  enabled: true
udp:
  enabled: true
configMap:
  name: configmap
  sslProtocols: TLSv1 TLSv1.1 TLSv1.2
  sslCiphers: ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-DSS-AES128-GCM-SHA256:kEDH+AESGCM:ECDHE-RSA-AES128-SHA256:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA:ECDHE-ECDSA-AES128-SHA:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA:ECDHE-ECDSA-AES256-SHA:DHE-RSA-AES128-SHA256:DHE-RSA-AES128-SHA:DHE-DSS-AES128-SHA256:DHE-RSA-AES256-SHA256:DHE-DSS-AES256-SHA:DHE-RSA-AES256-SHA:AES128-GCM-SHA256:AES256-GCM-SHA384:AES128-SHA256:AES256-SHA256:AES128-SHA:AES256-SHA:AES:CAMELLIA:DES-CBC3-SHA:!aNULL:!eNULL:!EXPORT:!DES:!RC4:!MD5:!PSK:!aECDH:!EDH-DSS-DES-CBC3-SHA:!EDH-RSA-DES-CBC3-SHA:!KRB5-DES-CBC3-SHA  # noqa
  proxyBodySize: 500m
  upstreamKeepaliveConnections: 64
"""

# K8S lb部署到的命名空间
K8S_LB_NAMESPACE = settings.BCS_SYSTEM_NAMESPACE

# release version prefix
RELEASE_VERSION_PREFIX = "(current-unchanged)"

try:
    from .constants_ext import *  # noqa
except ImportError as e:
    logger.debug('Load extension failed: %s', e)

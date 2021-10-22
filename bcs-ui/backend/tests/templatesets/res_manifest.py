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
NGINX_DEPLOYMENT1_JSON = {
    "apiVersion": "apps/v1beta2",
    "kind": "Deployment",
    "monitorLevel": "general",
    "webCache": {
        "volumes": [{"type": "emptyDir", "name": "", "source": ""}],
        "isUserConstraint": False,
        "remarkListCache": [{"key": "", "value": ""}],
        "labelListCache": [{"key": "app", "value": "nginx", "isSelector": True, "disabled": False}],
        "logLabelListCache": [{"key": "", "value": ""}],
        "isMetric": False,
        "metricIdList": [],
        "affinityYaml": "",
        "nodeSelectorList": [{"key": "", "value": ""}],
        "hostAliasesCache": [],
    },
    "customLogLabel": {},
    "metadata": {"name": "nginx-deployment-1"},
    "spec": {
        "minReadySeconds": 0,
        "replicas": 1,
        "strategy": {"type": "RollingUpdate", "rollingUpdate": {"maxUnavailable": 1, "maxSurge": 0}},
        "selector": {"matchLabels": {"app": "nginx"}},
        "template": {
            "metadata": {"labels": {"app": "nginx"}, "annotations": {}},
            "spec": {
                "restartPolicy": "Always",
                "terminationGracePeriodSeconds": 10,
                "nodeSelector": {},
                "affinity": {},
                "hostNetwork": 0,
                "dnsPolicy": "ClusterFirst",
                "volumes": [],
                "containers": [
                    {
                        "name": "container-1",
                        "webCache": {
                            "desc": "",
                            "imageName": "paas/k8stest/nginx",
                            "imageVersion": "",
                            "args_text": "",
                            "containerType": "container",
                            "livenessProbeType": "HTTP",
                            "readinessProbeType": "HTTP",
                            "logListCache": [{"value": ""}],
                            "env_list": [
                                {"type": "custom", "key": "eeee", "value": "{{test3}}.{{test7}}"},
                                {"type": "custom", "key": "fff", "value": "{{hieitest1}}"},
                            ],
                            "isImageCustomed": False,
                        },
                        "volumeMounts": [],
                        "image": "example.com:8443/paas/k8stest/nginx:{{image_version}}",
                        "imagePullPolicy": "IfNotPresent",
                        "ports": [{"id": 1570707798811, "containerPort": 80, "name": "http", "protocol": "TCP"}],
                        "command": "",
                        "args": "",
                        "env": [],
                        "workingDir": "",
                        "securityContext": {"privileged": False},
                        "resources": {
                            "limits": {"cpu": "", "memory": ""},
                            "requests": {"cpu": "", "memory": ""},
                        },
                        "livenessProbe": {
                            "httpGet": {"port": "", "path": "", "httpHeaders": []},
                            "tcpSocket": {"port": ""},
                            "exec": {"command": ""},
                            "initialDelaySeconds": 15,
                            "periodSeconds": 10,
                            "timeoutSeconds": 5,
                            "failureThreshold": 3,
                            "successThreshold": 1,
                        },
                        "readinessProbe": {
                            "httpGet": {"port": "", "path": "", "httpHeaders": []},
                            "tcpSocket": {"port": ""},
                            "exec": {"command": ""},
                            "initialDelaySeconds": 15,
                            "periodSeconds": 10,
                            "timeoutSeconds": 5,
                            "failureThreshold": 3,
                            "successThreshold": 1,
                        },
                        "lifecycle": {
                            "preStop": {"exec": {"command": ""}},
                            "postStart": {"exec": {"command": ""}},
                        },
                        "imageVersion": "{{image_version}}",
                        "logPathList": [],
                    }
                ],
                "initContainers": [],
            },
        },
    },
}

NGINX_DEPLOYMENT2_JSON = dict(NGINX_DEPLOYMENT1_JSON)
NGINX_DEPLOYMENT2_JSON["metadata"]["name"] = "nginx-deployment-2"

NGINX_SVC_JSON = {
    "apiVersion": "v1",
    "kind": "Service",
    "webCache": {"link_app": [], "link_labels": ["app:nginx"], "serviceIPs": ""},
    "metadata": {"name": "service-1", "labels": {}, "annotations": {}},
    "spec": {
        "type": "ClusterIP",
        "selector": {"app": "nginx"},
        "clusterIP": "",
        "ports": [
            {
                "name": "http",
                "port": 80,
                "protocol": "TCP",
                "targetPort": "http",
                "nodePort": "",
                "id": 1570707798811,
            }
        ],
    },
}

NGINX_DEPLOYMENT_YAML = """---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.7.9
        ports:
        - containerPort: 80
"""

CUSTOM_MANIFEST1_YAML = """---
apiVersion: v1
kind: Service
metadata:
  name: gw-node
  labels:
    app.kubernetes.io/instance: gw-node
spec:
  clusterIP: None
  ports:
    - name: http
      protocol: TCP
      port: 9100
      targetPort: 9100
---
apiVersion: v1
kind: Endpoints
metadata:
  name: gw-node
subsets:
  - addresses:
      - ip: 0.0.0.1
    ports:
      - name: http
        protocol: TCP
        port: 9100
"""

CUSTOM_MANIFEST2_YAML = """---
apiVersion: v1
kind: Pod
metadata:
  name: redis
spec:
  containers:
  - name: redis
    image: redis:5.0.4
    command:
      - redis-server
      - "/redis-master/redis.conf"
    env:
    - name: MASTER
      value: "true"
    ports:
    - containerPort: 6379
    resources:
      limits:
        cpu: "0.1"
    volumeMounts:
    - mountPath: /redis-master-data
      name: data
    - mountPath: /redis-master
      name: config
  volumes:
    - name: data
      emptyDir: {}
    - name: config
      configMap:
        name: example-redis-config
        items:
        - key: redis-config
          path: redis.conf
"""

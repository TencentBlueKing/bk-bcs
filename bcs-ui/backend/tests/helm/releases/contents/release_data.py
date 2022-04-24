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

# 解析后的release数据
release_data = {
    "name": "bk-redis",
    "info": {
        "first_deployed": "2021-05-21T14:47:53.959134+08:00",
        "last_deployed": "2021-05-21T14:47:53.959134+08:00",
        "deleted": "",
        "description": "Install complete",
        "status": "deployed",
        "notes": "1. Get the application URL by running these commands:\n  export POD_NAME=$(kubectl get pods --namespace default -l \"app=bk-redis1,release=bk-redis\" -o jsonpath=\"{.items[0].metadata.name}\")\n  echo \"Visit http://127.0.0.1:8080 to use your application\"\n  kubectl port-forward $POD_NAME 8080:80\n",  # noqa
    },
    "chart": {
        "metadata": {
            "name": "bk-redis1",
            "version": "0.1.46",
            "description": "A Helm chart for Kubernetie",
            "keywords": ["test"],
            "icon": "https://raw.githubusercontent.com/grafana/grafana/master/public/img/logo_transparent_400x.png",
            "apiVersion": "v2",
            "appVersion": "1.0",
            "annotations": {"test": "true", "test1": "false"},
        },
        "lock": None,
        "templates": [
            {
                "name": "templates/NOTES.txt",
            },
            {
                "name": "templates/_helpers.tpl",
            },
            {
                "name": "templates/deployment.yaml",
            },
            {"name": "templates/deployment1.yaml", "data": ""},
            {
                "name": "templates/hpa.yaml",
            },
            {
                "name": "templates/ingress.yaml",
            },
            {
                "name": "templates/service.yaml",
            },
        ],
        "values": {
            "affinity": {},
            "image": {"pullPolicy": "IfNotPresent", "repository": "demo.io/nginx", "tag": "latest"},
            "ingress": {
                "annotations": {},
                "enabled": False,
                "hosts": ["chart-example.local"],
                "path": "/",
                "tls": [],
            },
            "labels": [{"app-name": "test-db"}],
            "nodeSelector": {},
            "replicaCount": 1,
            "resources": {"limits": {"cpu": "100m", "memory": "128Mi"}},
            "service": {"port": 8080, "type": "ClusterIP"},
            "test": {"test1": "test1"},
            "tolerations": [],
        },
        "schema": None,
        "files": [
            {
                "name": ".helmignore",
            },
            {
                "name": "bcs-values/test.yaml",
            },
        ],
    },
    "manifest": "---\n# Source: bk-redis1/templates/service.yaml\napiVersion: v1\nkind: Service\nmetadata:\n  name: test12321\n  labels:\n    app: bk-redis1\n    chart: bk-redis1-0.1.46\n    release: bk-redis\n    heritage: Helm\nspec:\n  type: ClusterIP\n  ports:\n    - port: 8080\n      targetPort: http\n      protocol: TCP\n      name: http\n  selector:\n    app: bk-redis1\n    release: bk-redis\n---\n# Source: bk-redis1/templates/deployment.yaml\napiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: bk-redis-bk-redis1\n  labels:\n    app: bk-redis1\n    chart: bk-redis1-0.1.46\n    release: bk-redis\n    heritage: Helm\n    tet44: \"3342\"\n    test123: test1\n  annotations:\n    sidecar.tbusapp.io/inject: \"no\"\n    test123: \"yes\"\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: bk-redis1\n      release: bk-redis\n  template:\n    metadata:\n      labels:\n        app: bk-redis1\n        release: bk-redis\n    spec:\n      imagePullSecrets:\n      - name: \"test123123\"\n      containers:\n        - name: bk-redis1\n          image: \"demo.io/nginx:latest\"\n          imagePullPolicy: IfNotPresent\n          env:\n          - name: \"test\"\n            value: \"test\"\n          - name: \"test\"\n            value: \"test123\"\n          ports:\n            - name: http\n              containerPort: 80\n              protocol: TCP\n          livenessProbe:\n            httpGet:\n              path: /\n              port: http\n          readinessProbe:\n            httpGet:\n              path: /\n              port: http\n          resources:\n            limits:\n              cpu: 100m\n              memory: 128Mi\n---\n# Source: bk-redis1/templates/hpa.yaml\napiVersion: autoscaling/v2beta1\nkind: HorizontalPodAutoscaler\nmetadata:\n  name: bk-redis-bk-redis1\n  labels:\n    helm.sh/chart: testweb-0.2.6\n    app.kubernetes.io/name: testweb\n    app.kubernetes.io/instance: testweb-2005120733\n    app.kubernetes.io/version: v1.0.2.20200512\n    app.kubernetes.io/managed-by: Helm\n    io.tencent.paas.source_type: helm\nspec:\n  scaleTargetRef:\n    apiVersion: apps/v1\n    kind: Deployment\n    name: testweb-2005120733\n  minReplicas: 1\n  maxReplicas: 5\n  metrics:\n  - type: Resource\n    resource:\n      name: cpu\n      targetAverageUtilization: 80\n  - type: Resource\n    resource:\n      name: memory\n      targetAverageUtilization: 79\n",  # noqa
    "version": 1,
    "namespace": "default",
}


# 查询的命名空间下的release数据（仅包含用到数据），包含多个版本
release_list = [
    {
        "name": "test",
        "namespace": "default",
        "version": 1,
        "chart": {"metadata": {"name": "test", "version": "test"}, "values": {"replica": 2}},
        "config": {"replica": 1},
    },
    {
        "name": "test",
        "namespace": "default",
        "version": 2,
        "chart": {"metadata": {"name": "test", "version": "test"}, "values": {"replica": 2}},
        "config": {"replica": 1},
    },
    {
        "name": "test",
        "namespace": "default",
        "version": 3,
        "chart": {"metadata": {"name": "test", "version": "test"}, "values": {"replica": 2}},
        "config": {"replica": 3},
    },
]

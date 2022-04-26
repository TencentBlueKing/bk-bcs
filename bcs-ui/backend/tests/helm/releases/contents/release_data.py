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


secret_data = """SDRzSUFCdTZaMklDLzcxWGJXOFRPUkQrSzlZZUgrNTA3RnVhMHJJU0h5cEFCem9vVVZ2NFFsRGw3RHFKcWRkZTJkNjBvZXAvdnhsN1g5T0VBNlM3Z3RUdWVEd3Z6end6czNzZlNGcXlJQ1BCNGliVXJPQW1lRW9DTHBjS1pQZkJrbXRqcnd0V0NiVmxCYXBOa2trYUpzZmhKTDFLcDluMEpEcytpcDRmUDArUHBuOG1wMW1TNEhWQmYrRlN3UVN6WHQwL21senp5bklsVWZSV0drdUZJTGtxSzlSREZaRFkydUJwNXd1a1Vsbm1oR2xFL21LVzJEVWp0S29FenlrYUl4OHYzcEhGbHVoYVNpNVhlR3dZbWkycExFdzJsNFN3dTBwcFMyWWZYbDJmbjcxLy9lTEo3emYxZ3VWV2tCWFlxMVJoU0JnaWJLYWlPU01GVzlKYVdCSUtNZy9BMDRzV3lmU3BocHlvWVoxa0hwQlFrYTlHeVlyYTlZdDVjQjl4eTByek9ma1NsY3pTZ2xvYW9lV0hlZkNIQ3lWZks3RDZpUnR1eWRyYUtvdmpkSElTSmZBdnpVNlQwNFJZUldySVlLdHFQY3h6SHVEOU5uQk1LRndxZlV0MVFaNjBtUkUwQUZibU1uZ0E1UEkxMWRhVnZZM0ZQZXdTSkVXVU4weWJwalFRU1RSOXRxZG1aK1FORXlWeFpnazRKMzlETkZveXkxMzVidGoyVnVrQ2EvVTVnSnJaNEF0U0wvZVhNVmtEMldwNkc2MjRYZGNMeUZMRG9XWFNSbEN2ZUtYcGtrcmEvUzZCZFV6SFZiMEFER0plcm1LaFZ1cmFhaXFoVUJxdVhVK1Q1QzZxNUFyZDA0cC82cFBZVEx5c0dzalN5QkdUU3FDVXc5UTRQRnlvY0d4MTdmTEE1eFFGU3lvTUN4NFFTcUh5R3hESldnaW5BYVNsbnBhZmUwQTdjWHorNGVyMVpXVHZMTlpobjhMMW1va0tJb3RzSlE3cCtDWW9FWjR0TFg5QUxmVjZXRGhmNitEUW5YVkZ2MnNUR2trelk3NnJBOVhiOEp3MU9sanFEUlUxODVqUzVaSkxicmY0Z1BqeGtxNllPNmtBd3BtQ2ltN2RHRmllS3pzRFh4QStCcTVacGFBMWxONzZNVkNxaUt0WXJyaThjNldoSzVTN0FCeTJRUk9wZHpvdUxCNHpTUmZDRFNGWFRKQ3NsYkdlb283SElidWpPSUVpcURBVmpySFl5dWdsZGg2RlUvN2lTRUFYVERRMUIyYUZQU2JHaHNYQ2d5QlZ3UzVoVE9TUVJCc0ZaSVZ0L0ZMVkVwbVdPcEdCQnM4YnVBUXZ1ZlYvNWxYdHVKb2tKZm92QVFJUFJqbzVmYzg5R3h2b1BaN0s5VGgyUG9hN3JWeElMMFdOemZOMjVrQnFHSDdmVTl2LzRjNlVZTG9EelNkcThqVXJhYy8zSlJlN1hJK0F3Q1ZmU2FYWm1DQ0wzSVNlQ0RFNmFlbUJabUVpODJYVGJHRVl6dVZ2NU5LQmtKRnVGc1g3Q1RhWGZYdG5aSlBPNVEyWFJVWXV2YzVjdGhQT0RYeU1KU011eDhuUkpFV1JyNTA3SlRoV0J5Njl6TEZoSUEzOUdQU0h6ZHp2ajcxNHpUUUhTb0ljQitNY3hoTExuUXNzUTBhNklxQUk2OVQ2RDkyVEw1cVh3QldxWVJ2Tm5CeG5aU3V2dExJcVZ5SWpWeTlucmRBbjJLcVpobTdmeVc1UEFqOVFnWjBKTkM0Q3VERnhYNGxYbmU3ZVlyVG13MUZnLzB0VkhMck1UcWNaYk42am8rbkVyMUxTRXFSaENzb0dFNlFKeXZDQzVWUkhGdFlWQklqVGlNdXZBRGNhaytxUnFYbXdaUWFsSFJlYTVqZlE5M3VLVlZLYnI5OE5ZZGdQeElGczIxcTExb2JBNDQ4WVd6NWsreUNXWFJiNDQ0YjRERWJDSmNzMXM3M1pzQ255UEdod2dQOHRNbEJEMlBHVVMrRE5JSTV3aHhlRFVCcEhhRzYwQURJLzludkRvNWo4VHNuSWNLTU1GWm5jWk1QbmNjeGpvNFM0Q2JiLzhDY3VqbUJ3elR3WUFydjJoazNmL25UUXpacDVzYXV3ZHpxNHd2TU5rN0FZWjFvdDJJNUg5QVJ2MHRralk3RDRNaEkvRWorYVNaNHd0T0Qvc1lkbVJlNlk4Y3Z5a1cxWW5OQmlzRFozRC93S2hUTmNvRDgwOTlyWG81MkJWMXRsNENVQlhqbml6V1FCcmRZTnZ6ZEs4MjlZSzZCaGNkYm9NZjJya3hCM2EyVFdjVFA5a0VxM2JBR3pieEk5NjJabGROTzhmVFBqV3FUYmVhQjZTSW5qVjVmTWU4VndraVRINlNRNU9UbzZkR2ZUYjkwSUk0QnZQM2Zua0Q1c2VXakpJbHhzaHdPWXF3aGU5SE5jSlJXbEp2S2x2ZmFMY2ozZW5nNjhLN2NQTDlpeTJ3NTdWZzhlUEY0Ly9ZYmNuMlhKNWNWNEtKZjBycGNjT3dtem11ZStJR0d6enk4YVFyYmJ4ejlsNDZVTU5CeXY5RE1BRUFENWFMbmczOXh1YVJ2NXAreDZGdis3NlpQbjhQVTMrcUREMTgzdTY5YS9WcnNQM09EaEg5Sll1U010RUFBQQ=="""  # noqa

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
from django.conf import settings

# 镜像路径前缀
if settings.DEPOT_PREFIX:
    image_path_prefix = f'{settings.DEPOT_PREFIX}/public'
else:
    image_path_prefix = 'public/bcs'
image_prefix = f'{settings.DEVOPS_ARTIFACTORY_HOST}/{image_path_prefix}'

K8S_TEMPLATE = {
    "code": 0,
    "message": "OK",
    "data": {
        "K8sService": [
            {
                "id": 77,
                "name": "service-redis1",
                "deploy_tag_list": ["1527667491806501|K8sDeployment"],
                "service_tag": "1527670651988263",
                "config": {
                    "apiVersion": "v1",
                    "kind": "Service",
                    "webCache": {"link_app": [], "link_labels": ["app:redis"], "serviceIPs": ""},
                    "metadata": {"name": "service-redis1", "labels": {}, "annotations": {}},
                    "spec": {
                        "type": "ClusterIP",
                        "selector": {"app": "redis"},
                        "clusterIP": "",
                        "ports": [
                            {
                                "name": "port",
                                "port": 6379,
                                "protocol": "TCP",
                                "targetPort": "port",
                                "nodePort": "",
                                "id": 1527667131558,
                            }
                        ],
                    },
                },
            },
            {
                "id": 78,
                "name": "service-sts1",
                "deploy_tag_list": [],
                "service_tag": "1527670676238869",
                "config": {
                    "apiVersion": "v1",
                    "kind": "Service",
                    "webCache": {"link_app": [], "link_labels": [], "serviceIPs": ""},
                    "metadata": {"name": "service-sts1", "labels": {}, "annotations": {}},
                    "spec": {"type": "ClusterIP", "selector": {}, "clusterIP": "None", "ports": []},
                },
            },
            {
                "id": 79,
                "name": "service-nginx1",
                "service_tag": "1527670770986669",
                "deploy_tag_list": ["1527670584670192|K8sDeployment"],
                "config": {
                    "apiVersion": "v1",
                    "kind": "Service",
                    "webCache": {"link_app": [], "link_labels": ["app:nginx"], "serviceIPs": ""},
                    "metadata": {"name": "service-nginx1", "labels": {}, "annotations": {}},
                    "spec": {
                        "type": "NodePort",
                        "selector": {"app": "nginx"},
                        "clusterIP": "",
                        "ports": [
                            {
                                "id": 1527667508909,
                                "name": "nginx",
                                "port": 8080,
                                "protocol": "TCP",
                                "targetPort": "nginx",
                                "nodePort": "",
                            }
                        ],
                    },
                },
            },
        ],
        "K8sStatefulSet": [
            {
                "id": 12,
                "name": "statefulset-rumpetroll-v1",
                "desc": "",
                "deploy_tag": "1527671007581743",
                "config": {
                    "apiVersion": "apps/v1beta2",
                    "kind": "Deployment",
                    "webCache": {
                        "volumes": [{"type": "emptyDir", "name": "", "source": ""}],
                        "isUserConstraint": False,
                        "remarkListCache": [{"key": "", "value": ""}],
                        "labelListCache": [{"key": "app", "value": "rumpetroll", "isSelector": True}],
                        "logLabelListCache": [{"key": "", "value": ""}],
                        "isMetric": False,
                        "metricIdList": [],
                        "affinityYaml": "",
                    },
                    "customLogLabel": {},
                    "metadata": {"name": "statefulset-rumpetroll-v1"},
                    "spec": {
                        "replicas": 1,
                        "updateStrategy": {"type": "OnDelete", "rollingUpdate": {"partition": 0}},
                        "podManagementPolicy": "OrderedReady",
                        "volumeClaimTemplates": [
                            {
                                "metadata": {"name": ""},
                                "spec": {
                                    "accessModes": [],
                                    "storageClassName": "",
                                    "resources": {"requests": {"storage": 1}},
                                },
                            }
                        ],
                        "selector": {"matchLabels": {"app": "rumpetroll"}},
                        "template": {
                            "metadata": {"labels": {"app": "rumpetroll"}, "annotations": {}},
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
                                        "name": "container-rumpetroll-v1",
                                        "webCache": {
                                            "desc": "",
                                            # NOTE: imageName仅供前端匹配镜像使用，格式是镜像列表中name:value
                                            "imageName": f"{image_path_prefix}/k8s/pyrumpetroll:{image_path_prefix}/k8s/pyrumpetroll",  # noqa
                                            "imageVersion": "",
                                            "containerType": "container",
                                            "args_text": "",
                                            "livenessProbeType": "HTTP",
                                            "readinessProbeType": "HTTP",
                                            "logListCache": [{"value": ""}],
                                            "env_list": [
                                                {"type": "custom", "key": "DOMAIN", "value": "rumpetroll-game.bk.com"},
                                                {"type": "custom", "key": "MAX_CLIENT", "value": "2"},
                                                {"type": "custom", "key": "MAX_ROOM", "value": "100"},
                                                {"type": "custom", "key": "REDIS_HOST", "value": "service-redis1"},
                                                {"type": "custom", "key": "REDIS_PORT", "value": "6379"},
                                                {"type": "custom", "key": "REDIS_DB", "value": "0"},
                                                {"type": "custom", "key": "NUMPROCS", "value": "1"},
                                                {"type": "valueFrom", "key": "HOST", "value": "status.podIP"},
                                            ],
                                        },
                                        "volumeMounts": [],
                                        "image": f"{image_prefix}/k8s/pyrumpetroll:0.3",
                                        "imagePullPolicy": "IfNotPresent",
                                        "ports": [{"id": 1527670806610, "containerPort": 20000, "name": "port"}],
                                        "command": "",
                                        "args": "",
                                        "env": [],
                                        "envFrom": [],
                                        "resources": {
                                            "limits": {"cpu": "", "memory": ""},
                                            "requests": {"cpu": "", "memory": ""},
                                        },
                                        "livenessProbe": {
                                            "httpGet": {"port": "port", "path": "", "httpHeaders": []},
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
                                            "tcpSocket": {"port": "esdisc"},
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
                                        "imageVersion": "0.3",
                                        "logPathList": [],
                                    }
                                ],
                                "initContainers": [],
                            },
                        },
                    },
                    "monitorLevel": "general",
                },
                "service_tag": "1527670676238869",
            }
        ],
        "K8sDeployment": [
            {
                "id": 462,
                "deploy_tag": "1527667491806501",
                "name": "deploy-redis1",
                "desc": "",
                "config": {
                    "apiVersion": "apps/v1beta2",
                    "kind": "Deployment",
                    "webCache": {
                        "volumes": [{"type": "emptyDir", "name": "", "source": ""}],
                        "isUserConstraint": True,
                        "remarkListCache": [{"key": "", "value": ""}],
                        "labelListCache": [{"key": "app", "value": "redis", "isSelector": True}],
                        "logLabelListCache": [{"key": "", "value": ""}],
                        "isMetric": False,
                        "metricIdList": [],
                        "nodeSelectorList": [{"key": "app", "value": "redis"}],
                    },
                    "customLogLabel": {},
                    "metadata": {"name": "deploy-redis1"},
                    "spec": {
                        "minReadySeconds": 0,
                        "replicas": 1,
                        "strategy": {"type": "RollingUpdate", "rollingUpdate": {"maxUnavailable": 1, "maxSurge": 0}},
                        "selector": {"matchLabels": {"app": "redis"}},
                        "template": {
                            "metadata": {"labels": {"app": "redis"}, "annotations": {}},
                            "spec": {
                                "restartPolicy": "Always",
                                "terminationGracePeriodSeconds": 10,
                                "nodeSelector": {},
                                "affinity": {
                                    "podAntiAffinity": {
                                        "requiredDuringSchedulingIgnoredDuringExecution": [
                                            {
                                                "labelSelector": {
                                                    "matchExpressions": [
                                                        {"key": "app", "operator": "In", "values": ["redis"]}
                                                    ]
                                                },
                                                "topologyKey": "kubernetes.io/hostname",
                                            }
                                        ]
                                    }
                                },
                                "hostNetwork": 0,
                                "dnsPolicy": "ClusterFirst",
                                "volumes": [],
                                "containers": [
                                    {
                                        "name": "container-redis-default",
                                        "webCache": {
                                            "desc": "",
                                            # NOTE: imageName仅供前端匹配镜像使用，格式是镜像列表中name:value
                                            "imageName": f"{image_path_prefix}/k8s/redis:{image_path_prefix}/k8s/redis",  # noqa
                                            "imageVersion": "",
                                            "args_text": "",
                                            "containerType": "container",
                                            "livenessProbeType": "TCP",
                                            "readinessProbeType": "HTTP",
                                            "logListCache": [{"value": ""}],
                                            "env_list": [{"type": "custom", "key": "", "value": ""}],
                                        },
                                        "volumeMounts": [],
                                        "image": f"{image_prefix}/k8s/redis:1.0",
                                        "imagePullPolicy": "IfNotPresent",
                                        "ports": [{"id": 1527667131558, "containerPort": 6379, "name": "port"}],
                                        "command": "",
                                        "args": "",
                                        "env": [],
                                        "resources": {
                                            "limits": {"cpu": "", "memory": ""},
                                            "requests": {"cpu": "", "memory": ""},
                                        },
                                        "livenessProbe": {
                                            "httpGet": {"port": "", "path": "", "httpHeaders": []},
                                            "tcpSocket": {"port": "port"},
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
                                        "imageVersion": "1.0",
                                        "logPathList": [],
                                    }
                                ],
                                "initContainers": [],
                            },
                        },
                    },
                    "monitorLevel": "general",
                },
            },
            {
                "id": 463,
                "deploy_tag": "1527670584670192",
                "name": "deploy-nginx1",
                "desc": "",
                "config": {
                    "apiVersion": "apps/v1beta2",
                    "kind": "Deployment",
                    "webCache": {
                        "volumes": [{"type": "emptyDir", "name": "", "source": ""}],
                        "isUserConstraint": True,
                        "remarkListCache": [{"key": "", "value": ""}],
                        "labelListCache": [{"key": "app", "value": "nginx", "isSelector": True}],
                        "logLabelListCache": [{"key": "", "value": ""}],
                        "isMetric": False,
                        "metricIdList": [],
                    },
                    "customLogLabel": {},
                    "metadata": {"name": "deploy-nginx1"},
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
                                "affinity": {
                                    "podAntiAffinity": {
                                        "requiredDuringSchedulingIgnoredDuringExecution": [
                                            {
                                                "labelSelector": {
                                                    "matchExpressions": [
                                                        {"key": "app", "operator": "In", "values": ["nginx"]}
                                                    ]
                                                },
                                                "topologyKey": "kubernetes.io/hostname",
                                            }
                                        ]
                                    },
                                    "podAffinity": {
                                        "requiredDuringSchedulingIgnoredDuringExecution": [
                                            {
                                                "labelSelector": {
                                                    "matchExpressions": [
                                                        {"key": "app", "operator": "In", "values": ["redis"]}
                                                    ]
                                                },
                                                "topologyKey": "kubernetes.io/hostname",
                                            }
                                        ]
                                    },
                                },
                                "hostNetwork": 1,
                                "dnsPolicy": "ClusterFirstWithHostNet",
                                "volumes": [],
                                "containers": [
                                    {
                                        "name": "container-nginx-default",
                                        "webCache": {
                                            "desc": "",
                                            # NOTE: imageName仅供前端匹配镜像使用，格式是镜像列表中name:value
                                            "imageName": f"{image_path_prefix}/k8s/rumpetroll-openresty:{image_path_prefix}/k8s/rumpetroll-openresty",  # noqa
                                            "imageVersion": "",
                                            "args_text": "",
                                            "containerType": "container",
                                            "livenessProbeType": "TCP",
                                            "readinessProbeType": "HTTP",
                                            "logListCache": [{"value": ""}],
                                            "env_list": [
                                                {"type": "custom", "key": "DOMAIN", "value": "rumpetroll-game.bk.com"},
                                                {"type": "custom", "key": "MAX_CLIENT", "value": "2"},
                                                {"type": "custom", "key": "MAX_ROOM", "value": "100"},
                                                {"type": "custom", "key": "REDIS_HOST", "value": "service-redis1"},
                                                {
                                                    "type": "valueFrom",
                                                    "key": "NAMESPACE",
                                                    "value": "metadata.namespace",
                                                },
                                                {"type": "custom", "key": "REDIS_PORT", "value": "6379"},
                                                {"type": "custom", "key": "REDIS_DB", "value": "0"},
                                                {"type": "custom", "key": "PORT", "value": "80"},
                                            ],
                                        },
                                        "volumeMounts": [],
                                        "image": f"{image_prefix}/k8s/rumpetroll-openresty:0.51",  # noqa
                                        "imagePullPolicy": "IfNotPresent",
                                        "ports": [{"id": 1527667508909, "containerPort": 80, "name": "nginx"}],
                                        "command": "",
                                        "args": "",
                                        "env": [],
                                        "resources": {
                                            "limits": {"cpu": 300, "memory": 200},
                                            "requests": {"cpu": "", "memory": ""},
                                        },
                                        "livenessProbe": {
                                            "httpGet": {"port": "", "path": "", "httpHeaders": []},
                                            "tcpSocket": {"port": "nginx"},
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
                                        "imageVersion": "0.50",
                                        "logPathList": [],
                                    }
                                ],
                                "initContainers": [],
                            },
                        },
                    },
                    "monitorLevel": "general",
                },
            },
        ],
    },
}

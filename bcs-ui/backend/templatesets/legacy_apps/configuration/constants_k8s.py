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

k8s 配置模板的参数验证
DONE:
1. 亲和性后台校验：affinity
2. raymond 确认 job 中 replicas 和 parallelism 怎么配置

TODO:
"""
from ..instance.funutils import update_nested_dict
from .constants import FILE_DIR_PATTERN, NUM_VAR_PATTERN

# 资源名称
K8S_RES_NAME_PATTERN = "^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$"

# configmap/secret key 名称限制
KEY_NAME_PATTERN = "^[a-zA-Z{]{1}[a-zA-Z0-9-_.{}]{0,254}$"

# 端口名称限制
# PORT_NAME_PATTERN = "^[a-z{]{1}[a-z0-9-{}]{0,254}$"
# TODO 验证变量的情况
PORT_NAME_PATTERN = "^[a-zA-Z{]{1}[a-zA-Z0-9-{}_]{0,254}$"

# 挂载卷名称限制
VOLUMR_NAME_PATTERN = "^[a-zA-Z{]{1}[a-zA-Z0-9-_{}]{0,254}$"


# 亲和性验证
AFFINITY_MATCH_EXPRESSION_SCHNEA = {
    "type": "array",
    "items": {
        "type": "object",
        "required": ["key", "operator"],
        "properties": {
            "key": {"type": "string", "minLength": 1},
            "operator": {"type": "string", "enum": ["In", "NotIn", "Exists", "DoesNotExist", "Gt", "Lt"]},
            "values": {"type": "array", "items": {"type": "string", "minLength": 1}},
        },
        "additionalProperties": False,
    },
}

POD_AFFINITY_TERM_SCHNEA = {
    "type": "object",
    "properties": {
        "labelSelector": {"type": "object", "properties": {"matchExpressions": AFFINITY_MATCH_EXPRESSION_SCHNEA}},
        "namespaces": {"type": "array", "items": {"type": "string"}},
        "topologyKey": {"type": "string"},
    },
    "additionalProperties": False,
}

POD_AFFINITY_SCHNEA = {
    "type": "object",
    "properties": {
        "requiredDuringSchedulingIgnoredDuringExecution": {"type": "array", "items": POD_AFFINITY_TERM_SCHNEA},
        "preferredDuringSchedulingIgnoredDuringExecution": {
            "type": "array",
            "items": {
                "type": "object",
                "required": ["podAffinityTerm"],
                "properties": {
                    "weight": {
                        "oneOf": [
                            {"type": "string", "pattern": NUM_VAR_PATTERN},
                            {"type": "number", "minimum": 1, "maximum": 100},
                        ]
                    },
                    "podAffinityTerm": POD_AFFINITY_TERM_SCHNEA,
                },
            },
        },
    },
    "additionalProperties": False,
}

AFFINITY_SCHNEA = {
    "type": "object",
    "properties": {
        "nodeAffinity": {
            "type": "object",
            "properties": {
                "requiredDuringSchedulingIgnoredDuringExecution": {
                    "type": "object",
                    "required": ["nodeSelectorTerms"],
                    "properties": {
                        "nodeSelectorTerms": {
                            "type": "array",
                            "items": {
                                "type": "object",
                                "required": ["matchExpressions"],
                                "properties": {"matchExpressions": AFFINITY_MATCH_EXPRESSION_SCHNEA},
                            },
                        }
                    },
                },
                "preferredDuringSchedulingIgnoredDuringExecution": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "required": ["preference"],
                        "properties": {
                            "weight": {
                                "oneOf": [
                                    {"type": "string", "pattern": NUM_VAR_PATTERN},
                                    {"type": "number", "minimum": 1, "maximum": 100},
                                ]
                            },
                            "preference": {
                                "type": "object",
                                "required": ["matchExpressions"],
                                "properties": {"matchExpressions": AFFINITY_MATCH_EXPRESSION_SCHNEA},
                            },
                        },
                    },
                },
            },
            "additionalProperties": False,
        },
        "podAffinity": POD_AFFINITY_SCHNEA,
        "podAntiAffinity": POD_AFFINITY_SCHNEA,
    },
    "additionalProperties": False,
}

K8S_SECRET_SCHEM = {
    "type": "object",
    "required": ["metadata", "data"],
    "properties": {
        "metadata": {
            "type": "object",
            "required": ["name"],
            "properties": {"name": {"type": "string", "pattern": K8S_RES_NAME_PATTERN}},
        },
        "data": {
            "type": "object",
            "patternProperties": {KEY_NAME_PATTERN: {"type": "string"}},
            "additionalProperties": False,
        },
    },
}

K8S_CONFIGMAP_SCHEM = {
    "type": "object",
    "required": ["metadata", "data"],
    "properties": {
        "metadata": {
            "type": "object",
            "required": ["name"],
            "properties": {"name": {"type": "string", "pattern": K8S_RES_NAME_PATTERN}},
        },
        "data": {
            "type": "object",
            "patternProperties": {KEY_NAME_PATTERN: {"type": "string"}},
            "additionalProperties": False,
        },
    },
}

K8S_SERVICE_SCHEM = {
    "type": "object",
    "required": ["metadata", "spec"],
    "properties": {
        "metadata": {"type": "object", "required": ["name"], "properties": {"name": {"type": "string"}}},
        "spec": {
            "type": "object",
            "required": ["type", "clusterIP", "ports"],
            "properties": {
                "type": {"type": "string", "enum": ["ClusterIP", "NodePort", "LoadBalancer"]},
                "clusterIP": {"type": "string"},
                "ports": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "required": ["port", "protocol"],
                        "properties": {
                            "name": {
                                "oneOf": [
                                    {"type": "string", "pattern": "^$"},
                                    {"type": "string", "pattern": PORT_NAME_PATTERN},
                                ]
                            },
                            "port": {
                                "oneOf": [
                                    {"type": "string", "pattern": "^$"},
                                    {"type": "string", "pattern": NUM_VAR_PATTERN},
                                    {"type": "number", "minimum": 1, "maximum": 65535},
                                ]
                            },
                            "protocol": {"type": "string", "enmu": ["TCP", "UDP"]},
                            "targetPort": {
                                "anyof": [
                                    {"type": "number", "minimum": 1, "maximum": 65535},
                                    {"type": "string", "pattern": NUM_VAR_PATTERN},
                                    {"type": "string", "minLength": 1},
                                ]
                            },
                            "nodePort": {
                                "oneOf": [
                                    {"type": "string", "pattern": "^$"},
                                    {"type": "string", "pattern": NUM_VAR_PATTERN},
                                    {"type": "number", "minimum": 30000, "maximum": 32767},
                                ]
                            },
                        },
                    },
                },
            },
        },
    },
}

# 健康检查 & 就绪检查
K8S_CHECK_SCHEMA = {
    "type": "object",
    "required": [
        "initialDelaySeconds",
        "periodSeconds",
        "timeoutSeconds",
        "failureThreshold",
        "successThreshold",
    ],
    "properties": {
        "initialDelaySeconds": {
            "oneOf": [
                {"type": "string", "pattern": NUM_VAR_PATTERN},
                {"type": "number", "minimum": 0},
            ]
        },
        "periodSeconds": {
            "oneOf": [
                {"type": "string", "pattern": NUM_VAR_PATTERN},
                {"type": "number", "minimum": 1},
            ]
        },
        "timeoutSeconds": {
            "oneOf": [
                {"type": "string", "pattern": NUM_VAR_PATTERN},
                {"type": "number", "minimum": 1},
            ]
        },
        "failureThreshold": {
            "oneOf": [
                {"type": "string", "pattern": NUM_VAR_PATTERN},
                {"type": "number", "minimum": 1},
            ]
        },
        "successThreshold": {
            "oneOf": [
                {"type": "string", "pattern": NUM_VAR_PATTERN},
                {"type": "number", "minimum": 1},
            ]
        },
        "exec": {"type": "object", "properties": {"command": {"type": "string"}}},
        "tcpSocket": {"type": "object", "properties": {"port": {"oneOf": [{"type": "number"}, {"type": "string"}]}}},
        "httpGet": {
            "type": "object",
            "properties": {
                "port": {"oneOf": [{"type": "number"}, {"type": "string"}]},
                "path": {"type": "string"},
                "httpHeaders": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "properties": {
                            "name": {"type": "string"},
                            "value": {"type": "string"},
                        },
                    },
                },
            },
        },
    },
}

CONTAINER_SCHNEA = {
    "type": "array",
    "items": {
        "type": "object",
        "required": [
            "name",
            "image",
            "imagePullPolicy",
            "volumeMounts",
            "ports",
            "resources",
            "livenessProbe",
            "readinessProbe",
            "lifecycle",
        ],
        "properties": {
            "name": {"type": "string", "minLength": 1},
            "image": {"type": "string", "minLength": 1},
            "imagePullPolicy": {"type": "string", "enum": ["Always", "IfNotPresent", "Never"]},
            "volumeMounts": {
                "type": "array",
                "items": {
                    "type": "object",
                    "required": ["name", "mountPath", "readOnly"],
                    "properties": {
                        "name": {"type": "string", "pattern": VOLUMR_NAME_PATTERN},
                        "mountPath": {"type": "string", "pattern": FILE_DIR_PATTERN},
                        "readOnly": {"type": "boolean"},
                    },
                },
            },
            "ports": {
                "type": "array",
                "items": {
                    "type": "object",
                    "required": ["name", "containerPort"],
                    "properties": {
                        "name": {
                            "oneOf": [
                                {"type": "string", "pattern": "^$"},
                                {"type": "string", "pattern": PORT_NAME_PATTERN},
                            ]
                        },
                        "containerPort": {
                            "oneOf": [
                                {"type": "string", "pattern": "^$"},
                                {"type": "string", "pattern": NUM_VAR_PATTERN},
                                {"type": "number", "minimum": 1, "maximum": 65535},
                            ]
                        },
                    },
                },
            },
            "command": {"type": "string"},
            "args": {"type": "string"},
            # 环境变量前端统一存放在 webCache.env_list 中，有后台组装为 env & envFrom
            "env": {
                "type": "array",
                "items": {
                    "type": "object",
                    "required": ["name"],
                    "properties": {
                        "name": {"type": "string", "minLength": 1},
                        "value": {"type": "string"},
                        "valueFrom": {
                            "type": "object",
                            "properties": {
                                "fieldRef": {
                                    "type": "object",
                                    "required": ["fieldPath"],
                                    "properties": {"fieldPath": {"type": "string"}},
                                },
                                "configMapKeyRef": {
                                    "type": "object",
                                    "required": ["name", "key"],
                                    "properties": {
                                        "name": {"type": "string", "minLength": 1},
                                        "key": {"type": "string", "minLength": 1},
                                    },
                                },
                                "secretKeyRef": {
                                    "type": "object",
                                    "required": ["name", "key"],
                                    "properties": {
                                        "name": {"type": "string", "minLength": 1},
                                        "key": {"type": "string", "minLength": 1},
                                    },
                                },
                            },
                        },
                    },
                },
            },
            "envFrom": {
                "type": "array",
                "items": {
                    "type": "object",
                    "properties": {
                        "configMapRef": {
                            "type": "object",
                            "properties": {
                                "name": {"type": "string", "minLength": 1},
                            },
                        },
                        "secretRef": {
                            "type": "object",
                            "properties": {
                                "name": {"type": "string", "minLength": 1},
                            },
                        },
                    },
                },
            },
            "resources": {
                "type": "object",
                "properties": {
                    "limits": {
                        "type": "object",
                        "properties": {
                            "cpu": {
                                "oneOf": [
                                    {"type": "string", "pattern": "^$"},
                                    {"type": "string", "pattern": NUM_VAR_PATTERN},
                                    {"type": "number", "minimum": 0},
                                ]
                            },
                            "memory": {
                                "oneOf": [
                                    {"type": "string", "pattern": "^$"},
                                    {"type": "number", "minimum": 0},
                                    {"type": "string", "pattern": NUM_VAR_PATTERN},
                                ]
                            },
                        },
                    },
                    "requests": {
                        "type": "object",
                        "properties": {
                            "cpu": {
                                "oneOf": [
                                    {"type": "string", "pattern": "^$"},
                                    {"type": "number", "minimum": 0},
                                    {"type": "string", "pattern": NUM_VAR_PATTERN},
                                ]
                            },
                            "memory": {
                                "oneOf": [
                                    {"type": "string", "pattern": "^$"},
                                    {"type": "number", "minimum": 0},
                                    {"type": "string", "pattern": NUM_VAR_PATTERN},
                                ]
                            },
                        },
                    },
                },
            },
            "livenessProbe": K8S_CHECK_SCHEMA,
            "readinessProbe": K8S_CHECK_SCHEMA,
            "lifecycle": {
                "type": "object",
                "required": ["preStop", "postStart"],
                "properties": {
                    "preStop": {"type": "object", "required": ["exec"], "properties": {"command": {"type": "string"}}},
                    "postStart": {
                        "type": "object",
                        "required": ["exec"],
                        "properties": {"command": {"type": "string"}},
                    },
                },
            },
        },
    },
}

K8S_DEPLPYMENT_SCHNEA = {
    "type": "object",
    "required": ["metadata", "spec"],
    "properties": {
        "metadata": {
            "type": "object",
            "required": ["name"],
            "properties": {"name": {"type": "string", "pattern": K8S_RES_NAME_PATTERN}},
        },
        "spec": {
            "type": "object",
            "required": ["replicas", "strategy", "template"],
            "properties": {
                "replicas": {
                    "oneOf": [
                        {"type": "string", "pattern": NUM_VAR_PATTERN},
                        {"type": "number", "minimum": 0},
                    ]
                },
                "strategy": {
                    "type": "object",
                    "required": ["type"],
                    "properties": {
                        "type": {"type": "string", "enum": ["RollingUpdate", "Recreate"]},
                        "rollingUpdate": {"type": "object", "required": ["maxUnavailable", "maxSurge"]},
                    },
                },
                "template": {
                    "type": "object",
                    "required": ["metadata", "spec"],
                    "properties": {
                        "metadata": {
                            "type": "object",
                            "properties": {"lables": {"type": "object"}, "annotations": {"type": "object"}},
                        },
                        "spec": {
                            "type": "object",
                            "required": [
                                "restartPolicy",
                                "terminationGracePeriodSeconds",
                                "nodeSelector",
                                "hostNetwork",
                                "dnsPolicy",
                                "volumes",
                                "containers",
                            ],
                            "properties": {
                                "restartPolicy": {"type": "string", "enum": ["Always", "OnFailure", "Never"]},
                                "terminationGracePeriodSeconds": {
                                    "oneOf": [
                                        {"type": "string", "pattern": NUM_VAR_PATTERN},
                                        {"type": "number", "minimum": 0},
                                    ]
                                },
                                "nodeSelector": {"type": "object"},
                                "hostNetwork": {"oneOf": [{"type": "number"}, {"type": "string"}]},
                                "dnsPolicy": {
                                    "type": "string",
                                    "enum": ["ClusterFirst", "Default", "None", "ClusterFirstWithHostNet"],
                                },
                                "volumes": {
                                    "type": "array",
                                    "items": {
                                        "type": "object",
                                        "required": ["name"],
                                        "properties": {
                                            "name": {"type": "string", "pattern": VOLUMR_NAME_PATTERN},
                                            "hostPath": {
                                                "type": "object",
                                                "required": ["path"],
                                                "properties": {
                                                    "path": {"type": "string", "pattern": FILE_DIR_PATTERN},
                                                },
                                            },
                                            "emptyDir": {
                                                "type": "object",
                                            },
                                            "configMap": {
                                                "type": "object",
                                                "required": ["name"],
                                                "properties": {"name": {"type": "string", "minLength": 1}},
                                            },
                                            "secret": {
                                                "type": "object",
                                                "required": ["secretName"],
                                                "properties": {"secretName": {"type": "string", "minLength": 1}},
                                            },
                                            "persistentVolumeClaim": {
                                                "type": "object",
                                                "required": ["claimName"],
                                                "properties": {"claimName": {"type": "string", "minLength": 1}},
                                            },
                                        },
                                    },
                                },
                                "containers": CONTAINER_SCHNEA,
                                "initContainers": CONTAINER_SCHNEA,
                            },
                        },
                    },
                },
            },
        },
    },
}

# DS 与 Deployment 的差异项:滚动升级策略 中 选择 RollingUpdate 时，只可以选择 maxUnavailable
# "required": ["replicas", "strategy", "template"],
K8S_DAEMONSET_DIFF = {
    "properties": {
        "spec": {
            "required": ["updateStrategy", "template"],
            "properties": {"updateStrategy": {"properties": {"rollingUpdate": {"required": ["maxUnavailable"]}}}},
        }
    }
}
K8S_DAEMONSET_SCHNEA = update_nested_dict(K8S_DEPLPYMENT_SCHNEA, K8S_DAEMONSET_DIFF)

# Job 与 Deployment 的差异项: Pod 运行时设置
# TODO： raymond 确认 job 中 replicas 和 parallelism 怎么配置
K8S_JOB_DIFF = {
    "properties": {
        "spec": {
            "type": "object",
            "required": ["template", "completions", "parallelism", "backoffLimit", "activeDeadlineSeconds"],
            "properties": {
                "parallelism": {
                    "oneOf": [
                        {"type": "string", "pattern": NUM_VAR_PATTERN},
                        {"type": "number", "minimum": 0},
                    ]
                },
                "completions": {
                    "oneOf": [
                        {"type": "string", "pattern": NUM_VAR_PATTERN},
                        {"type": "number", "minimum": 0},
                    ]
                },
                "backoffLimit": {
                    "oneOf": [
                        {"type": "string", "pattern": NUM_VAR_PATTERN},
                        {"type": "number", "minimum": 0},
                    ]
                },
                "activeDeadlineSeconds": {
                    "oneOf": [
                        {"type": "string", "pattern": NUM_VAR_PATTERN},
                        {"type": "number", "minimum": 0},
                    ]
                },
            },
        }
    }
}
K8S_JOB_SCHNEA = update_nested_dict(K8S_DEPLPYMENT_SCHNEA, K8S_JOB_DIFF)

# statefulset 与 Deployment 的差异项
K8S_STATEFULSET_DIFF = {
    "properties": {
        "spec": {
            "required": ["template", "updateStrategy", "podManagementPolicy", "volumeClaimTemplates"],
            "properties": {
                "updateStrategy": {
                    "type": "object",
                    "required": ["type"],
                    "properties": {
                        "type": {"type": "string", "enum": ["OnDelete", "RollingUpdate"]},
                        "rollingUpdate": {
                            "type": "object",
                            "required": ["partition"],
                            "properties": {
                                "partition": {
                                    "oneOf": [
                                        {"type": "string", "pattern": NUM_VAR_PATTERN},
                                        {"type": "number", "minimum": 0},
                                    ]
                                }
                            },
                        },
                    },
                },
                "podManagementPolicy": {"type": "string", "enum": ["OrderedReady", "Parallel"]},
                "serviceName": {"type": "string", "minLength": 1},
                "volumeClaimTemplates": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "required": ["metadata", "spec"],
                        "properties": {
                            "metadata": {
                                "type": "object",
                                "required": ["name"],
                                "properties": {
                                    # "name": {"type": "string", "minLength": 1}
                                },
                            },
                            "spec": {
                                "type": "object",
                                "required": ["accessModes", "storageClassName", "resources"],
                                "properties": {
                                    #  "storageClassName": {"type": "string", "minLength": 1},
                                    "accessModes": {
                                        "type": "array",
                                        "items": {
                                            "type": "string",
                                            "enum": ["ReadWriteOnce", "ReadOnlyMany", "ReadWriteMany"],
                                        },
                                    },
                                    "resources": {
                                        "type": "object",
                                        "required": ["requests"],
                                        "properties": {
                                            "requests": {
                                                "type": "object",
                                                "required": ["storage"],
                                                "properties": {
                                                    "storage": {
                                                        "oneOf": [
                                                            {"type": "string", "pattern": NUM_VAR_PATTERN},
                                                            {"type": "number", "minimum": 0},
                                                        ]
                                                    }
                                                },
                                            }
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        }
    }
}
K8S_STATEFULSET_SCHNEA = update_nested_dict(K8S_DEPLPYMENT_SCHNEA, K8S_STATEFULSET_DIFF)

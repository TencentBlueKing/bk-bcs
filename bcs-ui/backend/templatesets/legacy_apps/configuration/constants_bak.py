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

配置模板的参数验证

TODO
1. RES_NAME_PATTERN/PORT_NAME_PATTERN 验证变量的情况
"""
from django.utils.translation import ugettext_lazy as _

# 资源名称 & 端口名称
# RES_NAME_PATTERN = "^[a-z{]{1}[a-z0-9-{}]{0,254}$"

# 验证变量的情况，并且支持BCS变量标识 $
RES_NAME_PATTERN = "^[a-zA-Z{\$]{1}[a-zA-Z0-9-{}_\$]{0,254}$"

# 挂载卷名称
VOL_NAME_PATTERN = "^[a-zA-Z{\$]{1}[a-zA-Z0-9-_{}\$]{0,254}$"

# configmap/secret key 名称限制
KEY_NAME_PATTERN = "^[a-zA-Z{\$]{1}[a-zA-Z0-9-_.{}\$]{0,254}$"

# 变量的格式
VARIABLE_PATTERN = "[A-Za-z][A-Za-z0-9-_]"
# 填写数字的地方，可以填写变量
NUM_VAR_PATTERN = "^{{%s*}}$" % VARIABLE_PATTERN
# 需要与 backend.templatesets.var_mgmt.serializers.py 中的说明保持一致
NUM_VAR_ERROR_MSG = _("只能包含字母、数字、中划线和下划线，且以字母开头")
# 文件目录正则
# FILE_DIR_PATTERN = "((?!\.)[\w\d\-\.\/~]+)+$"
FILE_DIR_PATTERN = "^((?!\.{\$)[\w\d\-\.\/~{}\$]+)+$"

SECRET_SCHEM = {
    "type": "object",
    "required": ["metadata", "datas"],
    "properties": {
        "metadata": {
            "type": "object",
            "required": ["name"],
            "properties": {"name": {"type": "string", "pattern": RES_NAME_PATTERN}},
        },
        "datas": {
            "type": "object",
            "patternProperties": {
                KEY_NAME_PATTERN: {
                    "type": "object",
                    "required": ["content"],
                    "properties": {"content": {"type": "string", "minLength": 1}},
                }
            },
            "additionalProperties": False,
        },
    },
}

CONFIGMAP_SCHEM = {
    "type": "object",
    "required": ["metadata", "datas"],
    "properties": {
        "metadata": {
            "type": "object",
            "required": ["name"],
            "properties": {"name": {"type": "string", "pattern": RES_NAME_PATTERN}},
        },
        "datas": {
            "type": "object",
            "patternProperties": {
                KEY_NAME_PATTERN: {
                    "anyOf": [
                        {
                            "type": "object",
                            "required": ["type", "content"],
                            "properties": {
                                "type": {"type": "string", "enum": ["file"]},
                                "content": {"type": "string", "minLength": 1},
                            },
                        },
                        {
                            "type": "object",
                            "required": ["type", "content"],
                            "properties": {
                                "type": {"type": "string", "enum": ["http"]},
                                "content": {"type": "string", "format": "uri"},
                            },
                        },
                    ]
                }
            },
            "additionalProperties": False,
        },
    },
}

SERVICE_SCHEM = {
    "type": "object",
    "required": ["metadata", "spec"],
    "properties": {
        "metadata": {
            "type": "object",
            "required": ["name"],
            "properties": {
                "name": {"type": "string", "pattern": RES_NAME_PATTERN},
                "lb_labels": {
                    "type": "object",
                    "required": ["BCSBALANCE"],
                    "properties": {"BCSBALANCE": {"type": "string", "enum": ["source", "roundrobin", "leastconn"]}},
                },
            },
        },
        "spec": {
            "type": "object",
            "required": ["type", "clusterIP", "ports"],
            "properties": {
                "type": {"type": "string", "enum": ["ClusterIP", "None"]},
                "clusterIP": {"oneOf": [{"type": "string"}, {"type": "array", "items": {"type": "string"}}]},
                "ports": {
                    "type": "array",
                    "items": {
                        "anyOf": [
                            {
                                "type": "object",
                                "required": ["protocol", "name", "servicePort"],
                                "properties": {
                                    "protocol": {"type": "string", "enmu": ["TCP", "UDP"]},
                                    "name": {"type": "string", "pattern": RES_NAME_PATTERN},
                                    "servicePort": {
                                        "oneOf": [
                                            {"type": "string", "pattern": NUM_VAR_PATTERN},
                                            {"type": "number", "minimum": 1, "maximum": 65535},
                                        ]
                                    },
                                },
                            },
                            {
                                "type": "object",
                                "required": ["protocol", "name", "servicePort", "domainName"],
                                "properties": {
                                    "protocol": {"type": "string", "enmu": ["HTTP"]},
                                    "name": {"type": "string", "pattern": RES_NAME_PATTERN},
                                    "servicePort": {
                                        "oneOf": [
                                            {"type": "string", "pattern": NUM_VAR_PATTERN},
                                            {"type": "number", "minimum": 1, "maximum": 65535},
                                        ]
                                    },
                                    "domainName": {"type": "string", "format": "hostname"},
                                },
                            },
                            {
                                "type": "object",
                                "required": ["protocol", "name", "servicePort", "domainName"],
                                "properties": {
                                    "protocol": {"type": "string", "enmu": ["HTTP", "TCP", "UDP"]},
                                    "name": {"type": "string", "pattern": "^$"},
                                    "servicePort": {"type": "string", "pattern": "^$"},
                                    "domainName": {"type": "string", "pattern": "^$"},
                                },
                            },
                        ]
                    },
                },
            },
        },
    },
}

DEPLPYMENT_SCHNEA = {
    "type": "object",
    "required": ["strategy"],
    "properties": {
        "strategy": {
            "type": "object",
            "required": ["type", "rollingupdate"],
            "properties": {
                "type": {"type": "string", "enum": ["RollingUpdate"]},
                "rollingupdate": {
                    "type": "object",
                    "required": ["maxUnavilable", "maxSurge", "upgradeDuration", "rollingOrder"],
                    "properties": {
                        "maxUnavilable": {
                            "oneOf": [
                                {"type": "string", "pattern": NUM_VAR_PATTERN},
                                {"type": "number", "minimum": 0},
                            ]
                        },
                        "maxSurge": {
                            "oneOf": [
                                {"type": "string", "pattern": NUM_VAR_PATTERN},
                                {"type": "number", "minimum": 0},
                            ]
                        },
                        "upgradeDuration": {
                            "oneOf": [
                                {"type": "string", "pattern": NUM_VAR_PATTERN},
                                {"type": "number", "minimum": 0},
                            ]
                        },
                        "rollingOrder": {"type": "string", "enum": ["CreateFirst", "DeleteFirst"]},
                    },
                },
            },
        }
    },
}

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
import copy
import json
from typing import Dict, List

from django.utils.translation import ugettext_lazy as _
from rest_framework.exceptions import ValidationError

from backend.components.base import ComponentAuth
from backend.components.paas_cc import PaaSCCClient

from ..models import NameSpaceVariable, Variable


def get_ns_id(access_token: str, project_id: str, cluster_id: str, namespace: str) -> int:
    """获取命名空间ID"""
    client = PaaSCCClient(ComponentAuth(access_token))
    data = client.get_cluster_namespace_list(project_id, cluster_id)
    # 匹配命名空间名称
    for ns in data.get("results") or []:
        if ns["name"] == namespace:
            return ns["id"]
    raise ValidationError(_("集群:{}下没有查询到命名空间:{}").format(cluster_id, namespace))


def get_var_data(ns_id: int) -> Dict[int, str]:
    """通过命名空间ID获取变量对应的值"""
    ns_vars = NameSpaceVariable.objects.filter(ns_id=ns_id).values("var_id", "data")
    return {var["var_id"]: var["data"] for var in ns_vars}


def get_var_key_and_name(var_ids: List[int]) -> Dict[int, Dict]:
    """通过变量id获取变量key和名称"""
    ns_vars = Variable.objects.filter(id__in=var_ids).values("id", "key", "name")
    return {var["id"]: {"key": var["key"], "name": var["name"]} for var in ns_vars}


def compose_data(var_id_data_map: Dict, var_id_key_name_map: Dict) -> List[Dict]:
    ns_var_list = []
    for var_id, key_name in var_id_key_name_map.items():
        ns_var_data = var_id_data_map.get(var_id)
        if not ns_var_data:
            continue
        ns_var = copy.deepcopy(key_name)
        ns_var.update({"id": var_id, "value": json.loads(ns_var_data)["value"]})
        ns_var_list.append(ns_var)
    return ns_var_list

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
import json
import logging
from django.conf import settings

from backend.components import paas_cc
from backend.helm.app.utils import yaml_dump, yaml_load
from backend.components.bcs import k8s
from backend.utils.local import local
from backend.templatesets.var_mgmt.constants import VariableScope
from backend.templatesets.var_mgmt.models import ClusterVariable, NameSpaceVariable, Variable

try:
    from backend.container_service.observability.datalog.utils import get_data_id_by_project_id
except ImportError:
    from backend.container_service.observability.datalog_ce.utils import get_data_id_by_project_id

logger = logging.getLogger(__name__)


def collect_system_variable(access_token, project_id, namespace_id):
    sys_variables = {}

    # 获取标准日志采集的dataid
    data_info = get_data_id_by_project_id(project_id)
    sys_variables['SYS_STANDARD_DATA_ID'] = data_info.get('standard_data_id')
    sys_variables['SYS_NON_STANDARD_DATA_ID'] = data_info.get('non_standard_data_id')

    resp = paas_cc.get_project(access_token, project_id)
    if resp.get('code') != 0:
        logger.error(
            "查询project的信息出错(project_id:{project_id}):{message}".format(
                project_id=project_id, message=resp.get('message')
            )
        )
    project_info = resp["data"]
    sys_variables["SYS_CC_APP_ID"] = project_info["cc_app_id"]
    sys_variables['SYS_PROJECT_KIND'] = project_info["kind"]
    sys_variables['SYS_PROJECT_CODE'] = project_info["english_name"]

    resp = paas_cc.get_namespace(access_token, project_id, namespace_id)
    if resp.get('code') != 0:
        logger.error(
            "查询命名空间的信息出错(namespace_id:{project_id}-{namespace_id}):{message}".format(
                namespace_id=namespace_id, project_id=project_id, message=resp.get('message')
            )
        )
    namespace_info = resp["data"]

    sys_variables["SYS_NAMESPACE"] = namespace_info["name"]
    sys_variables["SYS_CLUSTER_ID"] = namespace_info["cluster_id"]
    sys_variables["SYS_PROJECT_ID"] = namespace_info["project_id"]
    # SYS_JFROG_DOMAIN
    # SYS_NON_STANDARD_DATA_ID

    # 获取镜像地址
    jfrog_domain = paas_cc.get_jfrog_domain(access_token, project_id, sys_variables['SYS_CLUSTER_ID'])
    sys_variables['SYS_JFROG_DOMAIN'] = jfrog_domain

    return sys_variables


def get_namespace_variables(project_id, namespace_id):
    # 仅能拿到用户自定义的变量
    """
    project_var = [{
        'id': _v.id,
        'key': _v.key,
        'name': _v.name,
        'default_value': _v.get_default_value,
        'ns_values': ns_values
    }]
    ns_vars = NameSpaceVariable.get_ns_vars(namespace_id, project_id)
    ns_vars = [{
        'id': _v.id,
        'key': _v.key,
        'name': _v.name,
        'value': _ns_value if _ns_value else ''
    }]
    """
    project_var = NameSpaceVariable.get_project_ns_vars(project_id)

    namespace_vars = []
    for _var in project_var:
        _ns_values = _var['ns_values']
        _ns_value_ids = _ns_values.keys()
        namespace_vars.append(
            {
                'id': _var['id'],
                'key': _var['key'],
                'name': _var['name'],
                'value': _ns_values.get(namespace_id) if namespace_id in _ns_value_ids else _var['default_value'],
            }
        )

    ns_vars = NameSpaceVariable.get_ns_vars(namespace_id, project_id)
    namespace_vars += ns_vars
    variable = {item["key"]: item["value"] for item in namespace_vars}
    logger.info("get_namespace_variables %s:%s \n %s", project_id, namespace_id, json.dumps(variable))
    return variable


def get_cluster_variables(project_id, cluster_id):
    """查询集群下的变量"""
    cluster_vars = ClusterVariable.get_cluster_vars(cluster_id, project_id)
    return {info["key"]: info["value"] for info in cluster_vars}


def get_global_variables(project_id):
    vars = Variable.objects.filter(project_id=project_id, scope=VariableScope.GLOBAL.value)
    return {info.key: info.default_value for info in vars}


def get_bcs_variables(project_id, cluster_id, namespace_id):
    """获取变量值
    - 全局变量
    - 集群变量
    - 命名空间变量
    """
    project_code = paas_cc.get_project_code(local.request.user.token.access_token, project_id)
    namespace = paas_cc.get_namespace_name(local.request.user.token.access_token, project_id, namespace_id)

    client = k8s.K8SClient("", project_id, "", settings.DEFAULT_BCS_API_ENV)
    bcs_vars_map = client.render_vars(project_code, cluster_id, namespace, local.request.user.username, None)
    bcs_vars = {k: v['value'] for k, v in bcs_vars_map.items()}
    return bcs_vars


def get_ns_variables(project_id, cluster_id, ns_id, key_list):
    """获取多个变量值
    返回格式: {ns_id: ["key": "key", "name": "name"}]}
    如果key不存在, 返回空字符串(兼容老的逻辑)
    """
    if not key_list:
        return []

    ns_var_map = get_bcs_variables(project_id, cluster_id, ns_id)
    variable_list = []
    for key in key_list:
        value = ns_var_map.get(key, "")
        variable_list.append({"key": key, "name": key, "value": value})

    return variable_list


def get_multi_ns_variables(project_id, cluster_id, ns_list, key_list):
    """获取多个变量值
    返回格式: {ns_id: ["key": "key", "name": "name"}]}
    如果key不存在, 返回空字符串(兼容老的逻辑)
    """
    variable_dict = {}

    if not ns_list:
        return variable_dict

    keyList = ",".join(key_list)

    project_code = paas_cc.get_project_code(local.request.user.token.access_token, project_id)
    namespace_res = paas_cc.get_namespace_list(local.request.user.token.access_token, project_id, limit=10000)
    namespace_data = namespace_res.get("data", {}).get("results") or []
    namespace_dict = {str(i["id"]): i for i in namespace_data}

    client = k8s.K8SClient("", project_id, "", settings.DEFAULT_BCS_API_ENV)

    for ns_id in ns_list:
        ns_info = namespace_dict.get(ns_id)
        if not ns_info:
            continue
        ns_name = ns_info['name']
        ns_var_map = client.render_vars(
            project_code, ns_info['cluster_id'], ns_name, local.request.user.username, keyList
        )
        _v_list = []
        for _key in key_list:
            if _key in ns_var_map:
                _v_list.append(ns_var_map[_key])
            else:
                _v_list.append({"key": _key, "name": _key, "value": ""})
        variable_dict[ns_id] = _v_list
    return variable_dict


def merge_valuefile_with_bcs_variables(valuefile, bcs_variables, sys_variables):
    if sys_variables:
        bcs_variables.update(**sys_variables)
    valuefile = yaml_load(valuefile)
    if not valuefile:
        valuefile = {}
    valuefile["__BCS__"] = bcs_variables
    valuefile.setdefault("default", {})["__BCS__"] = bcs_variables
    return yaml_dump(valuefile)


def get_valuefile_with_bcs_variable_injected(access_token, project_id, namespace_id, valuefile, cluster_id):
    sys_variables = collect_system_variable(
        access_token=access_token, project_id=project_id, namespace_id=namespace_id
    )
    bcs_variables = get_bcs_variables(project_id, cluster_id, namespace_id)
    valuefile = merge_valuefile_with_bcs_variables(valuefile, bcs_variables, sys_variables)
    return valuefile

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

包含集群和节点相关的操作
- 集群安装
- 节点安装
- 节点删除
- 集群删除
"""
import json
import logging

from django.conf import settings

from backend.components.utils import request_factory

logger = logging.getLogger(__name__)

APIGW = 'bcs_ops'
ENV = 'prod'
API_HOST = '{APIGW_HOST}/api/apigw/{APIGW}/{STAG}'

FUNCTION_PATH_MAP = {
    'create_cluster': '/v1/install_cluster',
    'add_cluster_node': '/v1/add_node',
    'delete_cluster_node': '/v1/remove_node',
    'delete_cluster': '/v1/uninstall_cluster',
    'get_task_result': '/v1/query',
}


def get_request_url(function_name):
    api_host = API_HOST.format(APIGW_HOST=settings.APIGW_HOST, APIGW=APIGW, STAG=ENV)
    path = FUNCTION_PATH_MAP[function_name]
    return '{api_host}{path}'.format(api_host=api_host, path=path)


def get_headers(access_token, project_id):
    return {"X-BKAPI-AUTHORIZATION": json.dumps({"access_token": access_token, "project_id": project_id})}


def base_ops_request(function_name, access_token, project_id, method, json=None, params=None):
    url = get_request_url(function_name)
    http_request = request_factory(method)
    headers = get_headers(access_token, project_id)
    return http_request(url, json=json, params=params, headers=headers)


def create_cluster(
    access_token,
    project_id,
    k8s_mesos,
    cluster_id,
    master_ip_list,
    config,
    cc_module,
    control_ip,
    biz_id,
    username,
    websvr,
    platform,
):  # noqa
    """集群初始化流程"""
    data = {
        'type': k8s_mesos,
        'cluster_id': cluster_id,
        'master_ip_list': master_ip_list,
        'node_ip_list': [],
        'config': config,
        'cc_module_id': cc_module,
        'control_ip': control_ip,
        'biz_id': str(biz_id),
        'username': username,
        'websvr': websvr,
        "platform": platform,
    }
    return base_ops_request('create_cluster', access_token, project_id, 'post', json=data)


def add_cluster_node(
    access_token,
    project_id,
    k8s_mesos,
    cluster_id,
    master_ip_list,
    node_ip_list,
    config,
    control_ip,
    biz_id,
    username,
    cc_module,
    websvr,
):  # noqa
    """节点初始化流程"""
    data = {
        'type': k8s_mesos,
        'cluster_id': cluster_id,
        'master_ip_list': master_ip_list,
        'node_ip_list': node_ip_list,
        'config': config,
        'control_ip': control_ip,
        'biz_id': str(biz_id),
        'username': username,
        'cc_module_id': cc_module,
        'websvr': websvr,
    }
    return base_ops_request('add_cluster_node', access_token, project_id, 'post', json=data)


def delete_cluster_node(
    access_token,
    project_id,
    k8s_mesos,
    cluster_id,
    master_ip_list,
    node_ip_list,
    config,
    control_ip,
    biz_id,
    username,
    websvr,
):  # noqa
    """节点删除流程"""
    data = {
        'type': k8s_mesos,
        'cluster_id': cluster_id,
        'master_ip_list': master_ip_list,
        'node_ip_list': node_ip_list,
        'config': config,
        'control_ip': control_ip,
        'biz_id': str(biz_id),
        'username': username,
        'websvr': websvr,
    }
    return base_ops_request('delete_cluster_node', access_token, project_id, 'post', json=data)


def delete_cluster(
    access_token,
    project_id,
    k8s_mesos,
    cluster_id,
    master_ip_list,
    control_ip,
    biz_id,
    username,
    websvr,
    config=None,
    platform=None,
):  # noqa
    """集群删除流程"""
    data = {
        'type': k8s_mesos,
        'cluster_id': cluster_id,
        'master_ip_list': master_ip_list,
        'config': config,
        'control_ip': control_ip,
        'biz_id': str(biz_id),
        'username': username,
        'websvr': websvr,
        "platform": platform,
    }
    return base_ops_request('delete_cluster', access_token, project_id, 'post', json=data)


def get_task_result(access_token, project_id, task_id, biz_id, username):
    """查询任务状态"""
    params = {'task_id': task_id, 'biz_id': str(biz_id), 'username': username}
    return base_ops_request('get_task_result', access_token, project_id, 'get', params=params)


try:
    from .ops_ext import *  # noqa
except ImportError as e:
    logger.debug("Load ops_ext failed, %s", e)

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

from celery import shared_task

from backend.accounts import bcs_perm
from backend.components import paas_cc
from backend.utils import FancyDict
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes

from . import resources

logger = logging.getLogger(__name__)


def get_namespaces_by_bcs(access_token, project_id, project_kind, cluster_id):
    ns_client = resources.Namespace(access_token, project_id, project_kind)
    return ns_client.list(cluster_id)


def get_cluster_namespace_map(access_token, project_id):
    """获取项目下命名空间
    因为项目确定了容器编排类型，并且为减少多次请求的耗时，
    直接查询项目下的命名空间信息
    """
    # return data format: {'cluster_id': {'ns_name': 'ns_id'}}
    resp = paas_cc.get_namespace_list(access_token, project_id, desire_all_data=1)
    if resp.get('code') != ErrorCode.NoError:
        raise error_codes.APIError(f'get project namespace error, {resp.get("message")}')
    cluster_namespace_map = {}
    data = resp.get('data') or {}
    results = data.get('results') or []
    for info in results:
        cluster_id = info['cluster_id']
        if cluster_id in cluster_namespace_map:
            cluster_namespace_map[cluster_id][info['name']] = info['id']
        else:
            cluster_namespace_map[cluster_id] = {info['name']: info['id']}
    return cluster_namespace_map


def create_cc_namespace(access_token, project_id, cluster_id, namespace, creator):
    resp = paas_cc.create_namespace(access_token, project_id, cluster_id, namespace, None, creator, 'prod', True)
    if resp.get('code') != ErrorCode.NoError:
        raise error_codes.APIError(f'create namespace error, {resp.get("message")}')
    return resp['data']


def delete_cc_namespace(access_token, project_id, cluster_id, namespace_id):
    resp = paas_cc.delete_namespace(access_token, project_id, cluster_id, namespace_id)
    if resp.get('code') != ErrorCode.NoError:
        raise error_codes.APIError(f'delete namespace error, {resp.get("message")}')


def register_auth(request, project_id, cluster_id, ns_id, ns_name):
    perm = bcs_perm.Namespace(request, project_id, bcs_perm.NO_RES, cluster_id)
    perm.register(ns_id, f'{ns_name}({cluster_id})')


def delete_auth(request, project_id, ns_id):
    perm = bcs_perm.Namespace(request, project_id, ns_id)
    perm.delete()


def compose_request(access_token, username):
    """组装request，以便于使用auth api时使用"""
    return FancyDict(
        {
            'user': FancyDict({'username': username, 'token': FancyDict({'access_token': access_token})}),
        }
    )


@shared_task
def create_ns_flow(ns_params):
    ns_list = ns_params['add_ns_list']
    if not ns_list:
        return
    access_token = ns_params['access_token']
    project_id = ns_params['project_id']
    creator = ns_params['username']
    cluster_id = ns_params['cluster_id']
    request = compose_request(access_token, creator)
    for ns_name in ns_list:
        ns_info = create_cc_namespace(access_token, project_id, cluster_id, ns_name, creator)
        register_auth(request, project_id, cluster_id, ns_info['id'], ns_name)


@shared_task
def delete_ns_flow(ns_params):
    ns_list = ns_params['delete_ns_list']
    if not ns_list:
        return
    access_token = ns_params['access_token']
    project_id = ns_params['project_id']
    request = compose_request(access_token, ns_params['username'])
    for ns_name in ns_list:
        ns_id = ns_params['ns_id_map'][ns_name]
        delete_auth(request, project_id, ns_id)
        delete_cc_namespace(access_token, project_id, ns_params['cluster_id'], ns_id)


@shared_task
def sync_namespace(access_token, project_id, project_code, project_kind, cluster_id_list, username):
    """
    1. search bcs namespaces, example A
    2. search cc namespaces, example B
    3. diff cc and bcs namespaces:
       - if (bcs - cc), create cc namespace, jfrog account, secret and register auth,
       - if (cc - bcs), delete cc namespace records when project is k8s
    """
    # TODO: 先不校验权限
    cc_cluster_namespace_map = get_cluster_namespace_map(access_token, project_id)
    for cluster_id in cluster_id_list:
        bcs_ns_list = get_namespaces_by_bcs(access_token, project_id, project_kind, cluster_id)
        ns_id_map = cc_cluster_namespace_map.get(cluster_id) or {}
        cc_ns_list = ns_id_map.keys()

        # 只在线上存在的命名空间，需要进行创建操作
        add_ns_list = list(set(bcs_ns_list) - set(cc_ns_list))
        # create cc namespace record, create secret and register auth
        ns_params = {
            'access_token': access_token,
            'project_id': project_id,
            'project_code': project_code,
            'project_kind': project_kind,
            'cluster_id': cluster_id,
            'username': username,
        }
        create_ns_params = {'add_ns_list': add_ns_list, **ns_params}
        create_ns_flow.delay(create_ns_params)

        # 只在bcs cc上存在的命名空间，需要进行删除操作
        delete_ns_list = list(set(cc_ns_list) - set(bcs_ns_list))
        # delete cc namespace record and delete auth, when project_kind is k8s
        delete_ns_params = {'ns_id_map': ns_id_map, 'delete_ns_list': delete_ns_list, **ns_params}
        delete_ns_flow.delay(delete_ns_params)

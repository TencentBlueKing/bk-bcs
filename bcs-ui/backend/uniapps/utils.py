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
from backend.components import paas_cc
from backend.iam.permissions.resources.namespace_scoped import NamespaceScopedPermCtx, NamespaceScopedPermission
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes


def get_project_namespaces(access_token, project_id):
    """get all namespace from project"""
    ns_resp = paas_cc.get_namespace_list(access_token, project_id, desire_all_data=True)
    if ns_resp.get('code') != ErrorCode.NoError:
        raise error_codes.APIError(ns_resp.get('message'))
    data = ns_resp.get('data') or {}
    return data.get('results') or []


def get_cluster_namespaces(access_token, project_id, cluster_id):
    ns_resp = paas_cc.get_cluster_namespace_list(access_token, project_id, cluster_id, desire_all_data=True)
    if ns_resp.get('code') != ErrorCode.NoError:
        raise error_codes.APIError(f'get cluster namespace error, {ns_resp.get("message")}')
    data = ns_resp.get('data') or {}
    return data.get('results') or []


def get_namespace_id(access_token, project_id, cluster_ns_name, cluster_id=None):
    if cluster_id:
        namespace_list = get_cluster_namespaces(access_token, project_id, cluster_id)
    else:
        namespace_list = get_project_namespaces(access_token, project_id)
    for ns_info in namespace_list:
        cluster_ns_name_item = (ns_info['cluster_id'], ns_info['name'])
        if cluster_ns_name == cluster_ns_name_item:
            return ns_info['id']

    raise error_codes.ResNotFoundError(f'not found namespace: {cluster_ns_name}')


def can_use_namespace(request, project_id, cluster_id, namespace):
    perm_ctx = NamespaceScopedPermCtx(
        username=request.user.username, project_id=project_id, cluster_id=cluster_id, name=namespace
    )
    NamespaceScopedPermission().can_use(perm_ctx)


def can_use_namespaces(request, project_id, namespaces):
    """
    namespaces: [(cluster_id, ns_name)]
    """
    permission = NamespaceScopedPermission()
    for ns in namespaces:
        perm_ctx = NamespaceScopedPermCtx(
            username=request.user.username, project_id=project_id, cluster_id=ns[0], name=ns[1]
        )
        permission.can_use(perm_ctx)


def get_ns_id_map(access_token, project_id):
    namespace_data = get_project_namespaces(access_token, project_id)
    return {(ns['cluster_id'], ns['name']): ns['id'] for ns in namespace_data}

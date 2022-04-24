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
import base64
import json
import logging

from django.conf import settings

from backend.components import paas_cc
from backend.components.bcs.k8s import K8SClient
from backend.resources.namespace import NamespaceQuota
from backend.resources.namespace.constants import K8S_PLAT_NAMESPACE
from backend.templatesets.legacy_apps.instance.constants import K8S_IMAGE_SECRET_PRFIX
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes

logger = logging.getLogger(__name__)


def delete(access_token, project_id, cluster_id, ns_name):
    client = K8SClient(access_token, project_id, cluster_id, env=None)
    resp = client.delete_namespace(ns_name)
    if resp.get('code') == ErrorCode.NoError:
        return
    if 'not found' in resp.get('message'):
        return
    # k8s 删除资源配额
    quota_client = NamespaceQuota(access_token, project_id, cluster_id)
    quota_client.delete_namespace_quota(ns_name)

    raise error_codes.APIError(f'delete namespace error, name: {ns_name}, {resp.get("message")}')


def get_namespace(access_token, project_id, cluster_id):
    client = K8SClient(access_token, project_id, cluster_id, env=None)
    resp = client.get_namespace()
    if resp.get('code') != ErrorCode.NoError:
        raise error_codes.APIError(f'get namespace error, {resp.get("message")}')
    data = resp.get('data') or []
    namespace_list = []
    # 过滤掉 平台占用命名空间
    # TODO: 命名空间是否有状态不为Active的场景
    for info in data:
        resource_name = info['resourceName']
        if resource_name in K8S_PLAT_NAMESPACE:
            continue
        namespace_list.append(resource_name)
    return namespace_list

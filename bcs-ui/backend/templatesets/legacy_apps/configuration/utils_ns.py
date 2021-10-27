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

命名空间相关的方法
"""
from django.utils.translation import ugettext_lazy as _

from backend.accounts import bcs_perm
from backend.components import paas_cc
from backend.components.bcs.k8s import K8SClient
from backend.utils.basic import RequestClass

from .namespace.views import NamespaceBase


def register_default_ns(access_token, username, project_id, project_code, cluster_id):
    """注册默认的命名空间（针对k8s集群）
    1. 创建存储镜像账号的secret
    2. 将 default 命名空间注册到paas_cc 上
    project_code = request.project.english_name
    """
    # 组装创建ns的数据
    data = {'env_type': 'dev', 'name': 'default', 'cluster_id': cluster_id}
    ns_base = NamespaceBase()
    # 1. 创建存储镜像账号的secret
    client = K8SClient(access_token, project_id, data['cluster_id'], env=None)
    ns_base.create_jfrog_secret(client, access_token, project_id, project_code, data)

    # 2. 将 default 命名空间注册到paas_cc 上
    result = paas_cc.create_namespace(
        access_token, project_id, data['cluster_id'], data['name'], None, username, data['env_type']
    )
    if result.get('code') != 0:
        if 'Duplicate entry' in result.get('message', ''):
            message = _("创建失败，namespace名称已经在其他项目存在")
        else:
            message = result.get('message', '')
        return False, message

    # 注册资源到权限中心
    request = RequestClass(username=username, access_token=access_token, project_code=project_code)
    perm = bcs_perm.Namespace(request, project_id, bcs_perm.NO_RES, data['cluster_id'])
    perm.register(str(result['data']['id']), result['data']['name'])
    return True, _("命名空间[default]注册成功")

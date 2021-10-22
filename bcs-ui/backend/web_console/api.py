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

web_console暴露给其他服务使用的模块
"""
from django.utils.translation import ugettext_lazy as _

from backend.components.bcs.k8s import K8SClient

from .constants import WebConsoleMode
from .pod_life_cycle import K8SClient as _k8s_client
from .rest_api import utils


def exec_command(access_token: str, project_id: str, cluster_id: str, container_id: str, command: str) -> str:
    """
    在k8s容器中执行命令
    TODO 仅 backend.uniapps.application.common_views.query 使用，后续废弃
    """
    context = {}
    client = K8SClient(access_token, project_id, cluster_id, None)
    _context = utils.get_k8s_context(client, container_id)
    if not _context:
        raise ValueError(_("container_id不正确或者容器不是运行状态"))

    context.update(_context)

    try:
        bcs_context = utils.get_k8s_cluster_context(client, project_id, cluster_id)
    except Exception as error:
        raise ValueError(_('获取集群信息失败,{}').format(error))

    bcs_context = utils.get_k8s_admin_context(client, bcs_context, WebConsoleMode.INTERNEL.value)
    bcs_context['user_pod_name'] = context['pod_name']
    context.update(bcs_context)
    client = _k8s_client(context)
    result = client.exec_command(command)
    return result

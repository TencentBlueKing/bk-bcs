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
from typing import Dict, List

from django.conf import settings
from django.utils.translation import ugettext_lazy as _

from backend.components import cc
from backend.container_service.infras.hosts import perms as host_perms
from backend.utils.error_codes import error_codes
from backend.utils.exceptions import PermissionDeniedError
from backend.utils.funutils import convert_mappings

from .constants import CCHostKeyMappings

# 1表示gse agent正常
AGENT_NORMAL_STATUS = 1


def use_prometheus_source(request):
    """是否使用prometheus数据源"""
    if settings.DEFAULT_METRIC_SOURCE == 'prometheus':
        return True
    if request.project.project_code in settings.DEFAULT_METRIC_SOURCE_PROM_WLIST:
        return True
    return False


def can_use_hosts(bk_biz_id: int, username: str, host_ips: List):
    has_perm = host_perms.can_use_hosts(bk_biz_id, username, host_ips)
    if not has_perm:
        raise PermissionDeniedError(_("用户{}没有主机:{}的权限，请联系管理员在【配置平台】添加为业务运维人员角色").format(username, host_ips), "")


def get_cmdb_hosts(username: str, cc_app_id: int, host_property_filter: Dict) -> List:
    """
    根据指定条件获取 CMDB 主机信息，包含字段映射等转换
    """
    hosts = []
    for info in cc.HostQueryService(username, cc_app_id, host_property_filter=host_property_filter).fetch_all():
        convert_host = convert_mappings(CCHostKeyMappings, info)
        convert_host["agent"] = AGENT_NORMAL_STATUS
        hosts.append(convert_host)

    return hosts


def cluster_env_transfer(env_name, b2f=True):
    """tranfer name for frontend or cc"""
    if b2f:
        transfer_name = settings.CLUSTER_ENV_FOR_FRONT.get(env_name)
    else:
        transfer_name = settings.CLUSTER_ENV.get(env_name)
    if not transfer_name:
        raise error_codes.APIError(_("没有查询到集群所属环境"))
    return transfer_name


def get_nodes_repr(nodes: List[str]) -> str:
    """节点转换为字符串并截取为指定长度"""
    return ";".join(nodes)[:100]

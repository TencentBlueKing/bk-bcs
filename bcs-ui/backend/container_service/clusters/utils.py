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
from datetime import datetime
from typing import Dict, List

from django.conf import settings
from django.utils.translation import ugettext_lazy as _
from rest_framework.exceptions import ValidationError

from backend.components import cc
from backend.container_service.clusters.base import get_cluster_coes
from backend.container_service.clusters.base.constants import ClusterCOES
from backend.container_service.infras.hosts import perms as host_perms
from backend.utils.error_codes import error_codes
from backend.utils.exceptions import PermissionDeniedError
from backend.utils.funutils import convert_mappings

from .constants import CCHostKeyMappings

RoleNodeTag = 'N'
RoleMasterTag = 'M'
# 1表示gse agent正常
AGENT_NORMAL_STATUS = 1


def delete_node_labels_record(LabelModel, node_id_list, username):
    """删除数据库中关于节点标签的处理"""
    LabelModel.objects.filter(node_id__in=node_id_list, is_deleted=False).update(
        is_deleted=True, deleted_time=datetime.now(), updator=username, labels=json.dumps({})
    )


def gen_hostname_params(ip_list, cluster_id, is_master):
    return ["%s %s" % (ip, gen_hostname(ip, cluster_id, is_master)) for ip in ip_list]


def gen_hostname(ip, cluster_id, is_master):
    role = RoleMasterTag if is_master else RoleNodeTag
    ip_str = ip.replace('.', '-')
    host_name = "ip-%s-%s-%s" % (ip_str, role, cluster_id)
    return host_name.lower()


def cluster_env_transfer(env_name, b2f=True):
    """tranfer name for frontend or cc"""
    if b2f:
        transfer_name = settings.CLUSTER_ENV_FOR_FRONT.get(env_name)
    else:
        transfer_name = settings.CLUSTER_ENV.get(env_name)
    if not transfer_name:
        raise error_codes.APIError(_("没有查询到集群所属环境"))
    return transfer_name


def status_transfer(status, running_status_list, failed_status_list):
    """status display for frontend"""
    if status in running_status_list:
        return "running"
    elif status in failed_status_list:
        return "failed"
    return "success"


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


def use_tke(coes=None, access_token=None, project_id=None, cluster_id=None):
    """判断是否使用TKE"""
    if not (cluster_id or coes):
        raise ValidationError(_("集群ID或集群类型不能同时为空"))
    if not coes:
        coes = get_cluster_coes(access_token, project_id, cluster_id)

    if coes == ClusterCOES.TKE.value:
        return True
    return False


def get_ops_platform(request, coes=None, project_id=None, cluster_id=None):
    # 获取ops需要的平台类型，便于ops转发后面的标准运维
    # gcloud_v3_inner: 内部版标准运维v3, gcloud_v1_inner: 内部版标准运维v1, gcloud_v3_tke: tke流程
    access_token = request.user.token.access_token
    if use_tke(coes=coes, access_token=access_token, project_id=project_id, cluster_id=cluster_id):
        return 'gcloud_v3_tke'
    elif request.project.bg_id != getattr(settings, "IEG_ID", ""):
        return 'gcloud_v3_inner'
    else:
        return 'gcloud_v1_inner'

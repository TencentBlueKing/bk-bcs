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

from backend.components import bcs, paas_cc
from backend.container_service.projects.base.constants import LIMIT_FOR_ALL_DATA
from backend.templatesets.legacy_apps.configuration.models import Template
from backend.templatesets.legacy_apps.instance.constants import InsState
from backend.templatesets.legacy_apps.instance.models import InstanceConfig, VersionInstance
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes

logger = logging.getLogger(__name__)


def get_cluster_namespace(request, project_id, cluster_id):
    """获取集群下namespace的数量"""
    resp = paas_cc.get_cluster_namespace_list(
        request.user.token.access_token, project_id, cluster_id, limit=LIMIT_FOR_ALL_DATA
    )
    if resp.get("code") != ErrorCode.NoError:
        logger.error(u"获取命名空间数量异常，当前集群ID: %s, 详情: %s" % (cluster_id, resp.get("message")))
        return {}
    return resp.get("data") or {}


def get_used_namespace(project_id):
    """通过应用实例获取使用的命名空间"""
    # 通过project_id查询模板集信息
    all_tmpl = Template.objects.filter(project_id=project_id).values("id")
    tmpl_id_list = [info["id"] for info in all_tmpl]
    ns_id_info = VersionInstance.objects.filter(
        is_deleted=False, is_bcs_success=True, template_id__in=tmpl_id_list
    ).values("id")
    instance_id_list = [info["id"] for info in ns_id_info]
    inst_info = (
        InstanceConfig.objects.filter(
            is_deleted=False,
            is_bcs_success=True,
            instance_id__in=instance_id_list,
        )
        .exclude(ins_state=InsState.NO_INS.value)
        .values("namespace")
    )

    return [int(info["namespace"]) for info in inst_info]


def get_used_namespace_via_bcs(request, project_id, cluster_id, all_namespace_list):
    """通过bcs查询已经使用的命名空间"""
    client = bcs.k8s.K8SClient(request.user.token.access_token, project_id, cluster_id, None)
    used_ns_info = client.get_used_namespace()
    if used_ns_info.get("code") != ErrorCode.NoError:
        raise error_codes.APIError.f(used_ns_info.get("message"))
    return set(used_ns_info.get("data") or []) & set(all_namespace_list)


def get_namespace_metric(request, project_id, cluster_id):
    # NOTE: namespace的数量通过如下规则获取
    # 如果该命名空间下有应用实例，则认为被使用
    # 获取集群下命名空间信息
    namespace_map = get_cluster_namespace(request, project_id, cluster_id)
    # 获取总量
    namespace_total = namespace_map.get("count") or 0
    namespace_list = namespace_map.get("results") or []
    namespace_id_list = [info["id"] for info in namespace_list]
    namespace_name_list = [info["name"] for info in namespace_list]
    # TODO: 在wesley上线相关接口到正式环境后，再去掉下面两行逻辑
    # 获取使用的namespace
    used_namespace_id_list = get_used_namespace(project_id)
    # 取交集，获取总数量
    namespace_active = len(set(namespace_id_list) & set(used_namespace_id_list))
    try:
        used_namespace_list = get_used_namespace_via_bcs(request, project_id, cluster_id, namespace_name_list)
        namespace_active = len(used_namespace_list) or namespace_active
    except Exception:
        pass
    data = {'total': namespace_total, 'actived': namespace_active}
    return data

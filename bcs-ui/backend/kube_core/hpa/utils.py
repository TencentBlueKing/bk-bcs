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

from django.utils import timezone
from django.utils.translation import ugettext_lazy as _

from backend.bcs_web.audit_log import client as activity_client
from backend.container_service.clusters.base.models import CtxCluster
from backend.container_service.clusters.base.utils import get_cluster_type
from backend.container_service.clusters.constants import ClusterType
from backend.resources.constants import K8sResourceKind
from backend.resources.exceptions import DeleteResourceError
from backend.resources.hpa import client as hpa_client
from backend.resources.hpa.formatter import HPAFormatter
from backend.templatesets.legacy_apps.configuration.constants import K8sResourceName
from backend.templatesets.legacy_apps.instance.models import InstanceConfig
from backend.uniapps.application import constants as application_constants

logger = logging.getLogger(__name__)


def get_current_metrics_display(_current_metrics):
    """当前监控值前端显示"""
    current_metrics = []

    for name, value in _current_metrics.items():
        value["name"] = name.upper()
        # None 现在为-
        if value["current"] is None:
            value["current"] = "-"
        current_metrics.append(value)
    # 按CPU, Memory显示
    current_metrics = sorted(current_metrics, key=lambda x: x["name"])
    display = ", ".join(f'{metric["name"]}({metric["current"]}/{metric["target"]})' for metric in current_metrics)

    return display


def get_cluster_hpa_list(request, project_id, cluster_id, namespace=None):
    """获取基础hpa列表"""
    # 共享集群 HPA 不展示
    if get_cluster_type(cluster_id) == ClusterType.SHARED:
        return []

    project_code = request.project.english_name
    hpa_list = []

    try:
        ctx_cluster = CtxCluster.create(token=request.user.token.access_token, project_id=project_id, id=cluster_id)
        client = hpa_client.HPA(ctx_cluster)
        formatter = HPAFormatter(cluster_id, project_code)
        hpa_list = client.list(formatter=formatter, namespace=namespace)
    except Exception as error:
        logger.error("get hpa list error, %s", error)

    return hpa_list


def delete_hpa(request, project_id, cluster_id, ns_name, namespace_id, name):
    # 共享集群 HPA 不允许删除
    if get_cluster_type(cluster_id) == ClusterType.SHARED:
        raise DeleteResourceError(_("共享集群 HPA 不支持删除"))

    ctx_cluster = CtxCluster.create(token=request.user.token.access_token, project_id=project_id, id=cluster_id)
    client = hpa_client.HPA(ctx_cluster)
    try:
        client.delete_ignore_nonexistent(name=name, namespace=ns_name)
    except Exception as error:
        logger.error("delete hpa error, namespace: %s, name: %s, error: %s", ns_name, name, error)
        raise DeleteResourceError(_("删除HPA资源失败"))

    # 删除成功则更新状态
    InstanceConfig.objects.filter(namespace=namespace_id, category=K8sResourceName.K8sHPA.value, name=name).update(
        updator=request.user.username,
        oper_type=application_constants.DELETE_INSTANCE,
        deleted_time=timezone.now(),
        is_deleted=True,
        is_bcs_success=True,
    )


def get_deployment_hpa(request, project_id, cluster_id, ns_name, deployments):
    """通过deployment查询HPA关联信息"""
    hpa_list = get_cluster_hpa_list(request, project_id, cluster_id, namespace=ns_name)

    hpa_deployment_list = [i["ref_name"] for i in hpa_list if i["ref_kind"] == K8sResourceKind.Deployment.value]

    for deployment in deployments:
        if deployment["resourceName"] in hpa_deployment_list:
            deployment["hpa"] = True
        else:
            deployment["hpa"] = False

    return deployments


def activity_log(project_id, username, resource_name, description, status):
    """操作记录"""
    client = activity_client.ContextActivityLogClient(
        project_id=project_id, user=username, resource_type="hpa", resource=resource_name
    )
    if status is True:
        client.log_delete(activity_status="succeed", description=description)
    else:
        client.log_delete(activity_status="failed", description=description)

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

from django.http import HttpResponse
from prometheus_client import CollectorRegistry, generate_latest
from rest_framework import permissions
from rest_framework.renderers import BrowsableAPIRenderer

from backend.bcs_web.audit_log.constants import ActivityStatus, ActivityType, ResourceType
from backend.bcs_web.viewsets import SystemViewSet
from backend.container_service.observability.prom_exporter.exporter import Exporter
from backend.container_service.observability.prom_exporter.serializers import ExporterParamsSLZ
from backend.utils.renderers import BKAPIRenderer

logger = logging.getLogger(__name__)


class BaseViewSet(SystemViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)
    permission_classes = (permissions.IsAuthenticated,)
    content_type = "text/plain"


class NamespaceExporterViewSet(BaseViewSet):
    def export(self, request):
        """命名空间 export"""
        params = self.params_validate(ExporterParamsSLZ)
        try:
            registry = CollectorRegistry()
            exporter = Exporter(registry, [ResourceType.Namespace], params["start_at"], params["end_at"])
            exporter.add_summary_by_status([ActivityStatus.Failed.value, ActivityStatus.Error.value])
            return HttpResponse(
                generate_latest(registry),
                status=200,
                content_type=self.content_type,
            )
        except Exception as e:
            logger.error("Namespace metrics error: %s", e)
            return HttpResponse("# HELP Error occured", status=500, content_type=self.content_type)


class TemplateExporterViewSet(BaseViewSet):
    def export(self, request):
        """模板集exporter metric, 包含模板集创建失败次数、成功率"""
        params = self.params_validate(ExporterParamsSLZ)
        try:
            registry = CollectorRegistry()
            exporter = Exporter(registry, [ResourceType.Template], params["start_at"], params["end_at"])
            exporter.add_summary_by_status([ActivityStatus.Failed.value, ActivityStatus.Error.value])
            exporter.add_gauge_success_rate([ActivityType.Add])
            return HttpResponse(
                generate_latest(registry),
                status=200,
                content_type=self.content_type,
            )
        except Exception as e:
            logger.error("TemplateSet metrics error: %s", e)
            return HttpResponse("# HELP Error occured", status=500, content_type=self.content_type)


class HelmExporterViewSet(BaseViewSet):
    def export(self, request):
        """Helm相关exporter metric，包含部署成功率、更新成功率、回滚成功率"""
        params = self.params_validate(ExporterParamsSLZ)
        try:
            registry = CollectorRegistry()
            exporter = Exporter(registry, [ResourceType.HelmApp], params["start_at"], params["end_at"])
            # 创建成功率
            exporter.add_gauge_success_rate([ActivityType.Add])
            # 更新成功率
            exporter.add_gauge_success_rate([ActivityType.Modify])
            # 回滚成功率
            exporter.add_gauge_success_rate([ActivityType.Rollback])
            return HttpResponse(
                generate_latest(registry),
                status=200,
                content_type=self.content_type,
            )
        except Exception as e:
            logger.error("Helm metrics error: %s", e)
            return HttpResponse("# HELP Error occured", status=500, content_type=self.content_type)


class WorkloadsExporterViewSet(BaseViewSet):
    WORKLOADS = [
        ResourceType.Instance,
        ResourceType.Deployment,
        ResourceType.Job,
        ResourceType.DaemonSet,
        ResourceType.StatefulSet,
        ResourceType.CronJob,
    ]

    def export(self, request):
        """应用操作相关exporter metric，包含更新成功率和删除成功率"""
        params = self.params_validate(ExporterParamsSLZ)
        try:
            registry = CollectorRegistry()
            exporter = Exporter(registry, self.WORKLOADS, params["start_at"], params["end_at"])
            # 更新成功率
            exporter.add_gauge_success_rate([ActivityType.Modify], metric_name="worlad_update_success_rate")
            # 删除成功率
            exporter.add_gauge_success_rate([ActivityType.Delete], metric_name="worlad_delete_success_rate")
            return HttpResponse(
                generate_latest(registry),
                status=200,
                content_type=self.content_type,
            )
        except Exception as e:
            logger.error("Workload metrics error: %s", e)
            return HttpResponse("# HELP Error occured", status=500, content_type=self.content_type)


class ClusterToolsExporterViewSet(BaseViewSet):
    def export(self, request):
        """集群功能的 metric

        NOTE: 类型待补充后再调整，先设置为 crd 及 cobj
        """
        params = self.params_validate(ExporterParamsSLZ)
        try:
            registry = CollectorRegistry()
            exporter = Exporter(
                registry,
                [ResourceType.CRD, ResourceType.CustomObject],
                params["start_at"],
                params["end_at"],
            )
            # 安装成功率
            exporter.add_gauge_success_rate([ActivityType.Add], metric_name="clustertool_install_success_rate")
            # 更新成功率
            exporter.add_gauge_success_rate([ActivityType.Modify], metric_name="clustertool_update_success_rate")
            # 卸载成功率
            exporter.add_gauge_success_rate([ActivityType.Delete], metric_name="clustertool_uninstall_success_rate")
            return HttpResponse(
                generate_latest(registry),
                status=200,
                content_type=self.content_type,
            )
        except Exception as e:
            logger.error("ClusterTool metrics error: %s", e)
            return HttpResponse("# HELP Error occured", status=500, content_type=self.content_type)

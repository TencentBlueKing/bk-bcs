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

import semantic_version
from django.conf import settings
from django.utils.translation import ugettext_lazy as _
from rest_framework import viewsets
from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.response import Response

from backend.bcs_web.audit_log import client as activity_client
from backend.components.bcs import k8s
from backend.utils.error_codes import error_codes
from backend.utils.renderers import BKAPIRenderer

logger = logging.getLogger(__name__)


class PrometheusUpdateViewSet(viewsets.ViewSet):
    """更新 Prometheus 相关"""

    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def _get_version(self, image):
        version = image.rsplit(":", 1)[1]
        if version.startswith("v"):
            version = version[1:]
        return version

    def get(self, request, project_id, cluster_id):
        """是否需要更新 thano-sidecar 版本
        Deprecated 已经统一升级到 v2.5.0 版本
        """
        data = {"need_update": False, "update_tooltip": ""}
        return Response(data)

    def _activity_log(self, project_id, username, resource_name, description, status):
        """操作记录"""
        client = activity_client.ContextActivityLogClient(
            project_id=project_id, user=username, resource_type="metric", resource=resource_name
        )
        if status is True:
            client.log_delete(activity_status="succeed", description=description)
        else:
            client.log_delete(activity_status="failed", description=description)

    def update(self, request, project_id, cluster_id):
        access_token = request.user.token.access_token
        client = k8s.K8SClient(access_token, project_id, cluster_id, env=None)
        resp = client.get_prometheus("thanos", "po-prometheus-operator-prometheus")
        spec = resp.get("spec")
        if not spec:
            raise error_codes.APIError(_("Prometheus未安装, 请联系管理员解决"))

        need_update = False
        # 获取原来的值不变，覆盖更新
        for container in spec["containers"]:
            if container["name"] not in settings.PROMETHEUS_VERSIONS:
                continue

            image = settings.PROMETHEUS_VERSIONS[container["name"]]
            if semantic_version.Version(self._get_version(image)) <= semantic_version.Version(
                self._get_version(container["image"])
            ):
                continue

            need_update = True
            container["image"] = image

        if not need_update:
            raise error_codes.APIError(_("已经最新版本, 不需要升级"))

        patch_spec = {"spec": {"containers": spec["containers"]}}
        resp = client.update_prometheus("thanos", "po-prometheus-operator-prometheus", patch_spec)
        message = _("更新Metrics: 升级 thanos-sidecar 成功")
        self._activity_log(project_id, request.user.username, "update thanos-sidecar", message, True)
        return Response(resp)

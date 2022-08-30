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
from rest_framework import viewsets
from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.response import Response

from backend.helm.helm.models import ChartVersion
from backend.uniapps.network import constants, serializers
from backend.uniapps.network.constants import K8S_LB_CHART_NAME, K8S_LB_NAMESPACE
from backend.uniapps.network.views.charts.releases import HelmReleaseMixin
from backend.utils.renderers import BKAPIRenderer


class K8SIngressControllerViewSet(viewsets.ViewSet, HelmReleaseMixin):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)
    chart_name = K8S_LB_CHART_NAME
    namespace = K8S_LB_NAMESPACE
    public_repo_name = "public-repo"
    release_version_prefix = constants.RELEASE_VERSION_PREFIX

    def get_chart_versions(self, request, project_id):
        # 过滤公共仓库下面的lb chart名称
        chart_versions = (
            ChartVersion.objects.filter(
                name=self.chart_name,
                chart__repository__project_id=project_id,
                chart__repository__name=self.public_repo_name,
            )
            .order_by("-created")
            .values("version", "id")
        )
        # 获取release版本的version
        params = request.query_params
        cluster_id = params.get("cluster_id")
        # 查询是否release
        namespace = params.get("namespace") or self.namespace
        release = self.get_helm_release(cluster_id, name=self.chart_name, namespace=namespace)
        if not release:
            return Response(chart_versions)
        # id: -1, 表示此数据为组装数据，仅供前端展示匹配使用
        chart_versions = list(chart_versions)
        chart_versions.insert(
            0, {"version": f"{self.release_version_prefix} {release.get_current_version()}", "id": -1}
        )
        return Response(chart_versions)

    def get_version_detail(self, request, project_id):
        """获取指定版本chart信息，包含release的版本"""
        slz = serializers.ChartVersionSLZ(data=request.data)
        slz.is_valid(raise_exception=True)
        data = slz.validated_data
        version = data["version"]
        # 如果是release，查询release中对应的values信息
        if not version.startswith(self.release_version_prefix):
            chart_version = ChartVersion.objects.get(
                name=self.chart_name,
                version=version,
                chart__repository__project_id=project_id,
                chart__repository__name=self.public_repo_name,
            )
            version_detail = {"name": self.chart_name, "version": version, "files": chart_version.files}
            return Response(version_detail)
        # 获取release对应的values
        version_detail = {"name": self.chart_name, "version": version}
        namespace = data.get("namespace") or self.namespace
        cluster_id = data.get("cluster_id")
        release = self.get_helm_release(cluster_id, self.chart_name, namespace=namespace)
        if not release:
            return Response(version_detail)

        version_detail["files"] = release.release.chartVersionSnapshot.files
        # values 配置应该使用 release 的而不是特定版本 chart 默认的
        version_detail["files"][f"{self.chart_name}/values.yaml"] = release.release.valuefile

        return Response(version_detail)

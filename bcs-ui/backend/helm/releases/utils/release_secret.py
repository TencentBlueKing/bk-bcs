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
from typing import Dict, List

import yaml
from django.db.models.query import QuerySet

from backend.components.base import ComponentAuth
from backend.components.paas_cc import PaaSCCClient
from backend.container_service.clusters.base.models import CtxCluster
from backend.helm.app.models import App
from backend.helm.helm.bcs_variable import collect_system_variable
from backend.helm.helm.constants import TEMPORARY_APP_ID
from backend.helm.helm.models.chart import Chart, ChartRelease, ChartVersion, ChartVersionSnapshot
from backend.helm.helm.models.repo import Repository
from backend.resources.configs import secret
from backend.utils.basic import getitems

from .formatter import ReleaseSecretFormatter

logger = logging.getLogger(__name__)


def list_namespaced_releases(ctx_cluster: CtxCluster, namespace: str) -> List[Dict]:
    """查询namespace下的release
    NOTE: 为防止后续helm release对应的secret名称规则(sh.helm.release.v1.名称.v版本)变动，不直接根据secret名称进行过滤
    """
    client = secret.Secret(ctx_cluster)
    # 查询指定命名空间下的secrets
    return client.list(formatter=ReleaseSecretFormatter(), label_selector="owner=helm", namespace=namespace)


def get_release_detail(ctx_cluster: CtxCluster, namespace: str, release_name: str) -> Dict:
    """获取release详情"""
    release_list = list_namespaced_releases(ctx_cluster, namespace)
    release_list = [release for release in release_list if release.get("name") == release_name]
    if not release_list:
        logger.error(
            "not found release: [cluster_id: %s, namespace: %s, name: %s]", ctx_cluster.id, namespace, release_name
        )
        return {}
    # 通过release中的version对比，过滤到最新的 release data
    # NOTE: helm存储到secret中的release数据，每变动一次，增加一个secret，对应的revision就会增加一个，也就是最大的revision为当前release的存储数据
    return max(release_list, key=lambda item: item["version"])


def get_release_notes(ctx_cluster: CtxCluster, namespace: str, release_name: str) -> str:
    """查询release的notes"""
    release_detail = get_release_detail(ctx_cluster, namespace, release_name)
    return getitems(release_detail, "info.notes", "")


class RecordReleases:
    """记录release, 兼容前端使用

    NOTE: 因为后续会迁移到helm服务，针对已经存在的release不做处理
    """

    def __init__(self, ctx_cluster: CtxCluster, namespace: str):
        self.ctx_cluster = ctx_cluster
        self.namespace = namespace
        # 默认的操作者为admin
        self.operator = "admin"

    def record(self):
        """记录release数据"""
        # 获取chart
        chart_map = self._get_chart_map()
        releases = self._get_releases()
        namespace_id = self._get_namespace_id()
        system_variables = self._get_system_variables(namespace_id)
        # 记录到数据库
        for rel in releases:
            chart_name = rel["chart"]["metadata"]["name"]
            if chart_name not in chart_map:
                logger.info("chart: %s 不存在", chart_name)
                continue
            chart = chart_map[chart_name]
            chart_version = self._get_chart_version(rel["chart"]["metadata"]["version"], chart)
            if not chart_version:
                logger.info("chart version: %s 不存在", rel["chart"]["metadata"]["version"])
                continue
            values_content = self._get_values_content(rel)
            chart_release = self._record_chart_release(chart, chart_version, values_content)
            # 组装需要的数据
            rec = {
                "project_id": self.ctx_cluster.project_id,
                "cluster_id": self.ctx_cluster.id,
                "chart": chart,
                "release": chart_release,
                "namespace": self.namespace,
                "namespace_id": namespace_id,
                "name": rel["name"],
                "creator": self.operator,
                "updator": self.operator,
                "unique_ns": 0,  # 默认值
                "sys_variables": system_variables,
                "version": chart_version.version,
                "cmd_flags": '[]',
            }
            # 记录到db中
            app = App.objects.create(**rec)
            # 更新数据
            chart_release.app_id = app.id
            chart_release.save(update_fields=["app_id"])
            self._update_content(app)

    def _get_app_names(self) -> QuerySet:
        return App.objects.filter(cluster_id=self.ctx_cluster.id, namespace=self.namespace).values_list(
            "name", flat=True
        )

    def _get_chart_map(self) -> Dict:
        charts = Chart.objects.filter(repository__project_id=self.ctx_cluster.project_id)
        return {chart.name: chart for chart in charts}

    def _get_chart_version(self, version: str, chart: Chart) -> ChartVersion:
        return ChartVersion.objects.filter(name=chart.name, version=version, chart=chart).first()

    def _record_chart_release(self, chart: Chart, chart_version: ChartVersion, values_content: str) -> ChartRelease:
        snapshot = ChartVersionSnapshot.objects.make_snapshot(chart_version)
        repo = Repository.objects.filter(project_id=self.ctx_cluster.project_id).first()
        return ChartRelease.objects.create(
            repository=repo,
            chart=chart,
            chartVersionSnapshot=snapshot,
            answers=[],
            customs=[],
            valuefile=values_content,
            app_id=TEMPORARY_APP_ID,
            valuefile_name="values.yaml",  # 默认为values.yaml
        )

    def _get_releases(self):
        releases = list_namespaced_releases(self.ctx_cluster, self.namespace)
        app_names = self._get_app_names()
        release_map = {}
        for release in releases:
            # 跳过已经记录到db中的数据
            release_name = release["name"]
            if release_name in app_names:
                continue
            if release_name in release_map:
                release_map[release_name].append(release)
            else:
                release_map[release_name] = [release]
        # 通过release中的version对比，过滤到最新的 release，作为集群中运行的release信息
        return [max(release_map[name], key=lambda item: item["version"]) for name in release_map]

    def _get_values_content(self, release):
        """获取release中的values内容"""
        # release对应的values内容首先时config下的内容，如果config不存在，则是chart下的values内容
        value_content = release.get("config") or ""
        if not value_content:
            value_content = release["chart"]["values"]
        # json转换为yaml
        return yaml.dump(value_content)

    def _get_system_variables(self, namespace_id: int) -> Dict:
        """获取系统变量"""
        return collect_system_variable(
            access_token=self.ctx_cluster.context.auth.access_token,
            project_id=self.ctx_cluster.project_id,
            namespace_id=namespace_id,
        )

    def _get_namespace_id(self):
        """获取命名空间ID，用于权限相关验证"""
        client = PaaSCCClient(ComponentAuth(access_token=self.ctx_cluster.context.auth.access_token))
        namespaces = client.get_cluster_namespace_list(
            project_id=self.ctx_cluster.project_id, cluster_id=self.ctx_cluster.id
        )
        for ns in namespaces["results"]:
            if self.namespace != ns["name"]:
                continue
            return ns["id"]
        return -1

    def _update_content(self, app: App):
        content, _ = app.render_app(
            access_token=self.ctx_cluster.context.auth.access_token,
            username=app.updator,
        )
        release = app.release
        release.content = content
        release.save(update_fields=["content"])
        release.refresh_structure(app.namespace)

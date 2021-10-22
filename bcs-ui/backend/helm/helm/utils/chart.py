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
from dataclasses import dataclass
from typing import Dict, List, Set, Tuple

from django.db.models.query import QuerySet
from rest_framework.utils.serializer_helpers import ReturnList

from ..models.chart import Chart, ChartVersion
from ..models.repo import Repository
from ..serializers import ChartSLZ


@dataclass
class ChartList:
    project_id: str

    def get_chart_data(self) -> ReturnList:
        charts = Chart.objects.get_charts(self.project_id)
        slz = ChartSLZ(charts, many=True)
        charts_data = slz.data
        # 获取版本ID及仓库ID
        version_ids, repo_ids = self._get_version_and_repo_ids(charts_data)
        # 获取对应的版本及仓库信息
        versions = ChartVersion.objects.get_versions(version_ids)
        repos = Repository.objects.get_repos(repo_ids)
        return self._compose_data(charts_data, versions, repos)

    def _get_version_and_repo_ids(self, charts_data: ReturnList) -> Tuple[Set[int], Set[int]]:
        """获取版本和仓库的ID"""
        version_ids = []
        repo_ids = []
        for chart in charts_data:
            version_ids.append(chart["defaultChartVersion"])
            repo_ids.append(chart["repository"])
        return set(version_ids), set(repo_ids)

    def _compose_data(self, charts_data: ReturnList, versions: QuerySet, repos: QuerySet) -> ReturnList:
        version_id_map = {ver["id"]: ver for ver in versions}
        repo_id_map = {repo["id"]: repo for repo in repos}
        for chart in charts_data:
            chart["defaultChartVersion"] = version_id_map.get(chart["defaultChartVersion"]) or {}
            chart["repository"] = repo_id_map.get(chart["repository"]) or {}

        return charts_data

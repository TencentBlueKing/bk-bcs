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
import functools
from dataclasses import asdict, dataclass
from typing import List

from backend.components.bk_repo import BkRepoClient
from backend.helm.helm.models.chart import Chart, ChartVersion, ChartVersionSnapshot
from backend.helm.helm.models.repo import Repository
from backend.utils.async_run import async_run


def update_and_delete_chart_versions(project_id: str, project_code: str, chart: Chart, versions: List[str]):
    """更新或删除chart及versions
    TODO: 后续不存储DB, 可以废弃掉针对DB中存储的chart version相关记录的处理
    """
    chart_versions = ChartVersion.objects.filter(chart=chart, version__in=versions)
    version_digests = chart_versions.values_list("digest", flat=True)
    # 处理digest不变动的情况
    ChartVersionSnapshot.objects.filter(digest__in=version_digests).delete()
    chart_versions.delete()
    # 如果chart下没有版本了，需要删除当前chart；否则，更新默认版本为剩余版本中最新版本
    chart_all_versions = ChartVersion.objects.filter(chart=chart)
    if not chart_all_versions:
        chart.delete()
    else:
        chart.defaultChartVersion = chart_all_versions.order_by("-created")[0]
        chart.save(update_fields=["defaultChartVersion"])

    # 设置commit id为空，以防出现强制推送版本后，相同版本digest不变动的情况
    Repository.objects.filter(project_id=project_id, name=project_code).update(commit=None)


@dataclass
class ChartData:
    project_name: str
    repo_name: str
    chart_name: str


@dataclass
class RepoAuth:
    username: str
    password: str


def get_chart_version_list(chart_data: ChartData, repo_auth: RepoAuth) -> List[str]:
    """获取 chart 对应的版本列表"""
    client = BkRepoClient(username=repo_auth.username, password=repo_auth.password)
    chart_versions = client.get_chart_versions(**asdict(chart_data))
    # 如果不为列表，则返回为空
    if isinstance(chart_versions, list):
        return [info["version"] for info in chart_versions]
    return []


def delete_chart_version(chart_data: ChartData, client: BkRepoClient, version: str):
    """删除版本"""
    req_data = asdict(chart_data)
    req_data["version"] = version
    client.delete_chart_version(**req_data)


def batch_delete_chart_versions(chart_data: ChartData, repo_auth: RepoAuth, versions: List[str]):
    """批量删除chart版本"""
    # 组装并发任务
    client = BkRepoClient(username=repo_auth.username, password=repo_auth.password)
    delete_version = functools.partial(delete_chart_version, chart_data, client)
    tasks = [functools.partial(delete_version, version) for version in versions]
    async_run(tasks)

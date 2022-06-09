# -*- coding: utf-8 -*-
from typing import Dict, List, Optional, Tuple

from django.conf import settings
from rest_framework import serializers

from backend.components import bk_repo
from backend.helm.helm.models.chart import Chart, ChartVersion, ChartVersionSnapshot


def get_chart_version(
    project_name: str, repo_name: str, chart_name: str, version: str, username: str, password: str
) -> Dict:
    """调用接口获取仓库中指定版本的详情

    :param project_name: 项目名称
    :param repo_name: 仓库名称
    :param chart_name: 指定chart的名称，用于找到指定的chart
    :param version: 指定chart的版本
    :param username: 访问仓库的用户身份: 用户名
    :param password: 访问仓库的用户身份: 密码
    """
    client = bk_repo.BkRepoClient(username=username, password=password)
    return client.get_chart_version_detail(project_name, repo_name, chart_name, version)


def update_or_create_chart_version(chart: Chart, version_detail: Dict) -> ChartVersion:
    """更新或创建chart版本信息"""
    return ChartVersion.update_or_create_version(chart, version_detail)


def release_snapshot_to_version(chart_version_snapshot: ChartVersionSnapshot, chart: Chart) -> ChartVersion:
    """通过snapshot组装version数据"""
    return ChartVersion(id=0, chart=chart, keywords="chart version", **chart_version_snapshot.version_detail)


class VersionListSLZ(serializers.Serializer):
    name = serializers.CharField()
    version = serializers.CharField()
    created = serializers.CharField()
    urls = serializers.ListField(child=serializers.CharField())


class ReleaseVersionListSLZ(serializers.Serializer):
    name = serializers.CharField()
    version = serializers.CharField()
    created = serializers.CharField()


def sort_version_list(versions: List) -> List:
    versions.sort(key=lambda item: item["created"], reverse=True)
    return versions


def get_helm_project_and_repo_name(
    project_code: str, repo_name: Optional[str] = None, is_public_repo: Optional[bool] = None
) -> Tuple[str, str]:
    """获取项目及仓库名称

    :param project_code: BCS 项目编码
    :param repo_name: repo名称
    :param is_public_repo: 是否是公共仓库
    :returns: 返回项目名称和仓库名称
    """
    if is_public_repo or repo_name == settings.BCS_SHARED_CHART_REPO_NAME:
        return (settings.BK_REPO_SHARED_PROJECT_NAME, settings.BK_REPO_SHARED_CHART_DEPOT_NAME)

    # 针对项目下的chart仓库，项目名称和仓库名称一样
    return (project_code, project_code)

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
from typing import Dict, Optional

from backend.container_service.clusters.constants import ClusterType
from backend.container_service.clusters.featureflag.constants import UNSELECTED_CLUSTER, ViewMode
from backend.packages.blue_krill.data_types.enum import FeatureFlag, FeatureFlagField


class ClusterFeatureFlag(FeatureFlag):
    """对应左侧菜单功能，默认都开启"""

    CLUSTER = FeatureFlagField(name='CLUSTER', label='集群', default=True)
    OVERVIEW = FeatureFlagField(name='OVERVIEW', label='概览', default=True)
    NODE = FeatureFlagField(name='NODE', label='节点', default=True)
    NAMESPACE = FeatureFlagField(name='NAMESPACE', label='命名空间', default=True)
    TEMPLATESET = FeatureFlagField(name='TEMPLATESET', label='模板集', default=True)
    VARIABLE = FeatureFlagField(name='VARIABLE', label='变量管理', default=True)
    METRICS = FeatureFlagField(name='METRICS', label='Metric管理', default=True)
    HELM = FeatureFlagField(name='HELM', label='helm', default=True)
    WORKLOAD = FeatureFlagField(name='WORKLOAD', label='工作负载', default=True)
    NETWORK = FeatureFlagField(name='NETWORK', label='网络', default=True)
    CONFIGURATION = FeatureFlagField(name='CONFIGURATION', label='配置', default=True)
    RBAC = FeatureFlagField(name='RBAC', label='RBAC权限控制', default=True)
    REPO = FeatureFlagField(name='REPO', label='仓库', default=True)
    AUDIT = FeatureFlagField(name='AUDIT', label='操作审计', default=True)
    EVENT = FeatureFlagField(name='EVENT', label='事件查询', default=True)
    MONITOR = FeatureFlagField(name='MONITOR', label='监控中心', default=True)


class GlobalClusterFeatureFlag(ClusterFeatureFlag):
    """
    所有集群视图下关闭菜单：
    - 概览
    """

    OVERVIEW = FeatureFlagField(name='OVERVIEW', label='概览', default=False)


class SingleClusterFeatureFlag(ClusterFeatureFlag):
    """
    独立集群视图下关闭菜单：
    - 集群
    - 仓库
    - 操作审计
    """

    CLUSTER = FeatureFlagField(name='CLUSTER', label='集群', default=False)
    REPO = FeatureFlagField(name='REPO', label='仓库', default=False)
    AUDIT = FeatureFlagField(name='AUDIT', label='操作审计', default=False)


class DashboardClusterFeatureFlag(FeatureFlag):
    """ 资源视图特有 FeatureFlag """

    OVERVIEW = FeatureFlagField(name='OVERVIEW', label='集群总览', default=True)
    NODE = FeatureFlagField(name='NODE', label='节点', default=True)
    NAMESPACE = FeatureFlagField(name='NAMESPACE', label='命名空间', default=True)
    WORKLOAD = FeatureFlagField(name='WORKLOAD', label='工作负载', default=True)
    NETWORK = FeatureFlagField(name='NETWORK', label='网络', default=True)
    CONFIGURATION = FeatureFlagField(name='CONFIGURATION', label='配置', default=True)
    STORAGE = FeatureFlagField(name='STORAGE', label='存储', default=True)
    RBAC = FeatureFlagField(name='RBAC', label='RBAC', default=True)
    HPA = FeatureFlagField(name='HPA', label='HPA', default=True)
    CUSTOM_RESOURCE = FeatureFlagField(name='CUSTOM_RESOURCE', label='自定义资源', default=True)


class DashboardPublicClusterFeatureFlag(DashboardClusterFeatureFlag):
    """ 公共集群资源视图 FeatureFlag (部分禁用) """

    OVERVIEW = FeatureFlagField(name='OVERVIEW', label='集群总览', default=False)
    NODE = FeatureFlagField(name='NODE', label='节点', default=False)
    STORAGE = FeatureFlagField(name='STORAGE', label='存储', default=False)
    RBAC = FeatureFlagField(name='RBAC', label='RBAC', default=False)
    HPA = FeatureFlagField(name='HPA', label='HPA', default=False)
    CUSTOM_RESOURCE = FeatureFlagField(name='CUSTOM_RESOURCE', label='自定义资源', default=False)


class DashboardFederalClusterFeatureFlag(DashboardPublicClusterFeatureFlag):
    """ 联邦集群 FeatureFlag 暂时与公共集群一致 """


def get_cluster_feature_flags(
    cluster_id: str, feature_type: Optional[str], view_mode: Optional[str]
) -> Dict[str, bool]:
    """
    获取 feature_flags（页面菜单展示控制）

    :param cluster_id: 集群ID
    :param feature_type: 集群类型
    :param view_mode: 查看模式
    :return: feature_flags
    """
    # 资源视图类的走独立配置
    if view_mode == ViewMode.ResourceDashboard:
        # 公共集群必定是联邦集群，判断优先级较高
        if feature_type == ClusterType.PUBLIC:
            return DashboardPublicClusterFeatureFlag.get_default_flags()
        if feature_type == ClusterType.FEDERATION:
            return DashboardFederalClusterFeatureFlag.get_default_flags()
        return DashboardClusterFeatureFlag.get_default_flags()

    if cluster_id == UNSELECTED_CLUSTER:
        return GlobalClusterFeatureFlag.get_default_flags()

    if feature_type == ClusterType.SINGLE:
        return SingleClusterFeatureFlag.get_default_flags()

    return {}

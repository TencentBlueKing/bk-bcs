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
from backend.packages.blue_krill.data_types.enum import FeatureFlag, FeatureFlagField


class BaseFeatureFlag(FeatureFlag):
    """对应左侧菜单功能，默认都开启"""

    NAMESPACE = FeatureFlagField(name='NAMESPACE', label='命名空间', default=True)
    TEMPLATESET = FeatureFlagField(name='TEMPLATESET', label='模板集', default=True)
    VARIABLE = FeatureFlagField(name='VARIABLE', label='变量管理', default=True)
    HELM = FeatureFlagField(name='HELM', label='helm', default=True)


class GlobalClusterFeatureFlag(BaseFeatureFlag):
    """集群管理 - 全部集群"""

    CLUSTER = FeatureFlagField(name='CLUSTER', label='集群', default=True)
    NODE = FeatureFlagField(name='NODE', label='节点', default=True)
    WORKLOAD = FeatureFlagField(name='WORKLOAD', label='工作负载', default=True)
    NETWORK = FeatureFlagField(name='NETWORK', label='网络', default=True)
    CONFIGURATION = FeatureFlagField(name='CONFIGURATION', label='配置', default=True)
    REPO = FeatureFlagField(name='REPO', label='仓库', default=True)
    AUDIT = FeatureFlagField(name='AUDIT', label='操作审计', default=True)
    EVENT = FeatureFlagField(name='EVENT', label='事件查询', default=True)
    MONITOR = FeatureFlagField(name='MONITOR', label='监控中心', default=True)


class SingleClusterFeatureFlag(BaseFeatureFlag):
    """集群管理 - 单个独有集群"""

    OVERVIEW = FeatureFlagField(name='OVERVIEW', label='概览', default=True)
    NODE = FeatureFlagField(name='NODE', label='节点', default=True)
    WORKLOAD = FeatureFlagField(name='WORKLOAD', label='工作负载', default=True)
    NETWORK = FeatureFlagField(name='NETWORK', label='网络', default=True)
    CONFIGURATION = FeatureFlagField(name='CONFIGURATION', label='配置', default=True)
    EVENT = FeatureFlagField(name='EVENT', label='事件查询', default=True)
    MONITOR = FeatureFlagField(name='MONITOR', label='监控中心', default=True)


class SharedClusterFeatureFlag(BaseFeatureFlag):
    """集群管理 - 单个共享集群"""

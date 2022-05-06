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
    """资源视图公共 FeatureFlag"""

    NAMESPACE = FeatureFlagField(name='NAMESPACE', label='命名空间', default=True)
    WORKLOAD = FeatureFlagField(name='WORKLOAD', label='工作负载', default=True)
    NETWORK = FeatureFlagField(name='NETWORK', label='网络', default=True)
    CONFIGURATION = FeatureFlagField(name='CONFIGURATION', label='配置', default=True)
    CUSTOM_RESOURCE = FeatureFlagField(name='CUSTOM_RESOURCE', label='自定义资源', default=True)


class SingleClusterFeatureFlag(BaseFeatureFlag):
    """资源视图 - 独有集群 FeatureFlag"""

    OVERVIEW = FeatureFlagField(name='OVERVIEW', label='集群总览', default=True)
    NODE = FeatureFlagField(name='NODE', label='节点', default=True)
    STORAGE = FeatureFlagField(name='STORAGE', label='存储', default=True)
    RBAC = FeatureFlagField(name='RBAC', label='RBAC', default=True)
    HPA = FeatureFlagField(name='HPA', label='HPA', default=True)


class SharedClusterFeatureFlag(BaseFeatureFlag):
    """资源视图 - 共享集群 FeatureFlag"""

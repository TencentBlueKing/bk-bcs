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
from backend.container_service.clusters.featureflag.constants import UNSELECTED_CLUSTER_PLACEHOLDER, ViewMode

from . import cluster_mgr, dashboard


def get_cluster_feature_flags(
    cluster_id: str, cluster_type: Optional[ClusterType], view_mode: Optional[ViewMode]
) -> Dict[str, bool]:
    """
    获取 feature_flags（页面菜单展示控制）

    :param cluster_id: 集群ID
    :param cluster_type: 集群类型
    :param view_mode: 查看模式
    :return: feature_flags
    """
    if cluster_id == UNSELECTED_CLUSTER_PLACEHOLDER:
        return cluster_mgr.GlobalClusterFeatureFlag.get_default_flags()

    # 根据 view_mode 确定 feature_flag 模块
    feature_flag_module = {ViewMode.ResourceDashboard: dashboard, ViewMode.ClusterManagement: cluster_mgr}[view_mode]

    # 再根据集群类型获取相应 FeatureFlag 配置
    feature_flag = {
        ClusterType.SHARED: feature_flag_module.SharedClusterFeatureFlag,
        ClusterType.SINGLE: feature_flag_module.SingleClusterFeatureFlag,
    }[cluster_type]

    return feature_flag.get_default_flags()

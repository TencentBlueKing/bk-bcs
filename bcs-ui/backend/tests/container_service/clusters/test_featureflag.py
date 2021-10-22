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
import pytest

from backend.container_service.clusters.featureflag.constants import UNSELECTED_CLUSTER, ClusterFeatureType, ViewMode
from backend.container_service.clusters.featureflag.featflag import get_cluster_feature_flags


@pytest.mark.parametrize(
    'cluster_id, feature_type, view_mode, expected_flags',
    [
        (UNSELECTED_CLUSTER, None, ViewMode.ClusterManagement, {'CLUSTER': True, 'OVERVIEW': False, 'REPO': True}),
        (
            'BCS-K8S-40000',
            ClusterFeatureType.SINGLE,
            ViewMode.ClusterManagement,
            {'CLUSTER': False, 'OVERVIEW': True, 'REPO': False},
        ),
        (
            'BCS-K8S-40000',
            ClusterFeatureType.SINGLE,
            ViewMode.ResourceDashboard,
            {'NODE': True, 'WORKLOAD': True, 'CUSTOM_RESOURCE': True},
        ),
    ],
)
def test_get_cluster_feature_flags(cluster_id, feature_type: str, view_mode, expected_flags):
    feature_flags = get_cluster_feature_flags(cluster_id, feature_type, view_mode)
    for feature in expected_flags:
        assert feature_flags[feature] == expected_flags[feature]

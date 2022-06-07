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

from backend.container_service.clusters.constants import ClusterType
from backend.container_service.clusters.featureflag.constants import UNSELECTED_CLUSTER_PLACEHOLDER, ViewMode
from backend.container_service.clusters.featureflag.featflags import get_cluster_feature_flags
from backend.tests.conftest import TEST_SHARED_CLUSTER_ID


@pytest.mark.parametrize(
    'cluster_id, cluster_type, view_mode, expected_flags',
    [
        (
            UNSELECTED_CLUSTER_PLACEHOLDER,
            None,
            ViewMode.ClusterManagement,
            {
                'CLUSTER',
                'NAMESPACE',
                'TEMPLATESET',
                'VARIABLE',
                'HELM',
                'NODE',
                'WORKLOAD',
                'NETWORK',
                'CONFIGURATION',
                'REPO',
                'AUDIT',
                'EVENT',
                'MONITOR',
            },
        ),
        (
            'BCS-K8S-40000',
            ClusterType.SINGLE,
            ViewMode.ClusterManagement,
            {
                'OVERVIEW',
                'NODE',
                'NAMESPACE',
                'TEMPLATESET',
                'VARIABLE',
                'HELM',
                'WORKLOAD',
                'NETWORK',
                'CONFIGURATION',
                'EVENT',
                'MONITOR',
            },
        ),
        (
            TEST_SHARED_CLUSTER_ID,
            ClusterType.SHARED,
            ViewMode.ClusterManagement,
            {
                'NAMESPACE',
                'TEMPLATESET',
                'VARIABLE',
                'HELM',
            },
        ),
        (
            'BCS-K8S-40000',
            ClusterType.SINGLE,
            ViewMode.ResourceDashboard,
            {
                'OVERVIEW',
                'NODE',
                'NAMESPACE',
                'WORKLOAD',
                'NETWORK',
                'CONFIGURATION',
                'STORAGE',
                'RBAC',
                'HPA',
                'CUSTOM_RESOURCE',
            },
        ),
        (
            TEST_SHARED_CLUSTER_ID,
            ClusterType.SHARED,
            ViewMode.ResourceDashboard,
            {'NAMESPACE', 'WORKLOAD', 'NETWORK', 'CONFIGURATION', 'CUSTOM_RESOURCE'},
        ),
    ],
)
def test_get_cluster_feature_flags(cluster_id, cluster_type, view_mode, expected_flags):
    feature_flags = get_cluster_feature_flags(cluster_id, cluster_type, view_mode)
    # 选择单集群或不选择集群时候，ieod 集群管理会额外注入 featureflags，这两种情况只检查 expected_flags 是否为子集即可
    if view_mode == ViewMode.ClusterManagement and cluster_type in [None, ClusterType.SHARED, ClusterType.SINGLE]:
        assert not expected_flags - feature_flags.keys()
    else:
        assert feature_flags.keys() == expected_flags

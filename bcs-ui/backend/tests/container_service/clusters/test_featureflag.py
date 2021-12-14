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
                'CLUSTER': True,
                'NAMESPACE': True,
                'TEMPLATESET': True,
                'VARIABLE': True,
                'METRICS': True,
                'HELM': True,
                'NODE': True,
                'WORKLOAD': True,
                'NETWORK': True,
                'CONFIGURATION': True,
                'REPO': True,
                'AUDIT': True,
                'EVENT': True,
                'MONITOR': True,
            },
        ),
        (
            'BCS-K8S-40000',
            ClusterType.SINGLE,
            ViewMode.ClusterManagement,
            {
                'OVERVIEW': True,
                'NODE': True,
                'NAMESPACE': True,
                'TEMPLATESET': True,
                'VARIABLE': True,
                'METRICS': True,
                'HELM': True,
                'WORKLOAD': True,
                'NETWORK': True,
                'CONFIGURATION': True,
                'EVENT': True,
                'MONITOR': True,
            },
        ),
        (
            TEST_SHARED_CLUSTER_ID,
            ClusterType.SHARED,
            ViewMode.ClusterManagement,
            {
                'NAMESPACE': True,
                'TEMPLATESET': True,
                'VARIABLE': True,
                'METRICS': True,
                'HELM': True,
            },
        ),
        (
            'BCS-K8S-40000',
            ClusterType.SINGLE,
            ViewMode.ResourceDashboard,
            {
                'OVERVIEW': True,
                'NODE': True,
                'NAMESPACE': True,
                'WORKLOAD': True,
                'NETWORK': True,
                'CONFIGURATION': True,
                'STORAGE': True,
                'RBAC': True,
                'HPA': True,
                'CUSTOM_RESOURCE': True,
            },
        ),
        (
            TEST_SHARED_CLUSTER_ID,
            ClusterType.SHARED,
            ViewMode.ResourceDashboard,
            {
                'NAMESPACE': True,
                'WORKLOAD': True,
                'WORKLOAD_DAEMONSET': False,
                'NETWORK': True,
                'CONFIGURATION': True,
                'STORAGE': True,
                'STORAGE_PV': False,
                'STORAGE_SC': False,
                'RBAC': True,
                'HPA': True,
                'CUSTOM_RESOURCE': True,
            },
        ),
    ],
)
def test_get_cluster_feature_flags(cluster_id, cluster_type, view_mode, expected_flags):
    feature_flags = get_cluster_feature_flags(cluster_id, cluster_type, view_mode)
    for key in expected_flags:
        assert expected_flags[key] == feature_flags[key]

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
import copy

import mock
from django.conf import settings

from backend.container_service.clusters.base.utils import append_shared_clusters

fake_shared_clusters = [{"cluster_id": "BCS-K8S-00000"}, {"cluster_id": "BCS-K8S-00001"}]
fake_project_clusters = [{"cluster_id": "BCS-K8S-00001"}]


class TestAddSharedClusters:
    @mock.patch('backend.container_service.clusters.base.utils.cm.get_shared_clusters', return_value=[])
    def test_for_null_shared_cluster(self, get_shared_clusters):
        project_clusters = []
        assert append_shared_clusters(project_clusters) == project_clusters

        project_clusters = copy.deepcopy(fake_project_clusters)
        assert append_shared_clusters(project_clusters) == project_clusters

    @mock.patch(
        'backend.container_service.clusters.base.utils.cm.get_shared_clusters',
        return_value=fake_shared_clusters,
    )
    def test_for_existed_shared_cluster(self, get_shared_clusters):
        project_clusters = []
        assert append_shared_clusters(project_clusters) == fake_shared_clusters

        project_clusters = copy.deepcopy(fake_project_clusters)
        project_clusters = append_shared_clusters(project_clusters)
        assert len(project_clusters) == 2

        project_clusters = copy.deepcopy(fake_shared_clusters)
        assert len(project_clusters) == 2
        assert append_shared_clusters(project_clusters) == project_clusters

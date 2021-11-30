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

import pytest
from django.conf import settings

from backend.container_service.clusters.base.utils import add_public_clusters

fake_cluster_id = "BCS-K8S-00000"
fake_public_clusters = [{"cluster_id": fake_cluster_id}]
fake_project_clusters = [{"cluster_id": "BCS-K8S-00001"}]


class TestAddPublicClusters:
    def test_for_null_public_cluster(self):
        settings.PUBLIC_CLUSTERS = []
        project_clusters = []
        assert add_public_clusters(project_clusters) == project_clusters

        project_clusters = copy.deepcopy(fake_project_clusters)
        assert add_public_clusters(project_clusters) == project_clusters

    def test_for_existed_public_cluster(self):
        settings.PUBLIC_CLUSTERS = copy.deepcopy(fake_public_clusters)
        # 项目集群为空
        project_clusters = []
        assert add_public_clusters(project_clusters) == fake_public_clusters
        # 公共集群包含在项目集群中
        project_clusters = copy.deepcopy(fake_public_clusters)
        assert len(project_clusters) == 1
        assert add_public_clusters(project_clusters) == project_clusters
        # 公共集群不在项目集群中
        project_clusters = copy.deepcopy(fake_project_clusters)
        project_clusters = add_public_clusters(project_clusters)
        assert len(project_clusters) == 2

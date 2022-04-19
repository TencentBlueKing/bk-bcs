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
from requests_mock import ANY

from backend.components.cluster_manager import ClusterManagerClient


class TestClusterManagerClient:
    def test_get_nodes(self, cluster_id, request_user, requests_mock):
        expected_data = [{"innerIP": "127.0.0.1"}]
        requests_mock.get(ANY, json={"code": 0, "data": expected_data})

        client = ClusterManagerClient(request_user.token.access_token)
        data = client.get_nodes(cluster_id)
        assert data == expected_data

    def test_get_shared_clusters(self, cluster_id, requests_mock):
        requests_mock.get(ANY, json={"code": 0, "data": [{"clusterID": cluster_id}]})

        data = ClusterManagerClient().get_shared_clusters()
        assert isinstance(data, list)
        assert cluster_id in [info["clusterID"] for info in data]

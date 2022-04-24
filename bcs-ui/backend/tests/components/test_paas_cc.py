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
import mock
from requests_mock import ANY

from backend.components.base import ComponentAuth
from backend.components.paas_cc import PaaSCCClient, UpdateNodesData


class TestPaaSCCClient:
    @mock.patch("backend.components.paas_cc.get_shared_clusters", return_value=[])
    def test_get_cluster_simple(self, get_shared_clusters, project_id, cluster_id, requests_mock):
        requests_mock.get(ANY, json={'foo': 'bar'})

        client = PaaSCCClient(ComponentAuth('token'))
        resp = client.get_cluster(project_id, cluster_id)
        assert resp == {'foo': 'bar'}
        assert requests_mock.called

    def test_get_cluster_by_id(self, cluster_id, requests_mock):
        requests_mock.get(ANY, json={'code': 0, 'data': {'cluster_id': cluster_id}})

        client = PaaSCCClient(ComponentAuth('token'))
        resp = client.get_cluster_by_id(cluster_id)
        assert resp == {'cluster_id': cluster_id}
        assert requests_mock.called

    def test_update_cluster(self, project_id, cluster_id, requests_mock):
        requests_mock.put(
            ANY, json={"code": 0, "data": {"cluster_id": cluster_id, "project_id": project_id, "status": "normal"}}
        )
        client = PaaSCCClient(ComponentAuth('token'))
        resp = client.update_cluster(project_id, cluster_id, {"status": "normal"})
        assert resp == {"cluster_id": cluster_id, "project_id": project_id, "status": "normal"}
        assert requests_mock.called

    def test_delete_cluster(self, project_id, cluster_id, requests_mock):
        requests_mock.delete(ANY, json={"code": 0, "data": None})
        client = PaaSCCClient(ComponentAuth('token'))
        resp = client.delete_cluster(project_id, cluster_id)
        assert resp is None
        assert requests_mock.called
        assert requests_mock.request_history[0].method == "DELETE"

    def test_update_node_list(self, project_id, cluster_id, requests_mock):
        requests_mock.patch(
            ANY,
            json={
                "code": 0,
                "data": [
                    {"inner_ip": "127.0.0.1", "cluster_id": cluster_id, "project_id": project_id, "status": "normal"}
                ],
            },
        )
        client = PaaSCCClient(ComponentAuth('token'))
        resp = client.update_node_list(
            project_id, cluster_id, [UpdateNodesData(inner_ip="127.0.0.1", status="normal")]
        )
        assert resp == [
            {"inner_ip": "127.0.0.1", "cluster_id": cluster_id, "project_id": project_id, "status": "normal"}
        ]
        assert requests_mock.called
        assert requests_mock.request_history[0].method == "PATCH"

    def test_get_node_list(self, project_id, cluster_id, requests_mock):
        requests_mock.get(ANY, json={"code": 0, "data": {"count": 1, "results": [{"inner_ip": "127.0.0.1"}]}})
        client = PaaSCCClient(ComponentAuth("token"))
        resp = client.get_node_list(project_id, cluster_id)

        assert resp == {"count": 1, "results": [{"inner_ip": "127.0.0.1"}]}
        assert "desire_all_data" in requests_mock.last_request.qs

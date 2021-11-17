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

from backend.resources.namespace import CtxCluster, Namespace, getitems

pytestmark = pytest.mark.django_db

COMMON_CLUSTER_ID = 'BCS-K8S-95001'

COMMON_CLUSTER_NS_NAME = 'unittest-proj-test-ns'

COMMON_CLUSTER_NS_MANIFEST = {
    "apiVersion": "v1",
    "kind": "Namespace",
    "metadata": {"annotations": {"io.tencent.paas.projectcode": "unittest-proj"}, "name": COMMON_CLUSTER_NS_NAME},
}


class TestNamespace:
    def test_list(self, api_client, project_id, cluster_id):
        """ 测试获取资源列表接口 """
        response = api_client.get(f'/api/dashboard/projects/{project_id}/clusters/{cluster_id}/namespaces/')
        assert response.json()['code'] == 0

    def test_list_common_cluster_ns(self, api_client, project_id):
        """ 获取公共集群中项目拥有的命名空间 """
        client = Namespace(CtxCluster.create(token='access_token', id=COMMON_CLUSTER_ID, project_id=project_id))
        client.create(body=COMMON_CLUSTER_NS_MANIFEST)
        response = api_client.get(f'/api/dashboard/projects/{project_id}/clusters/{COMMON_CLUSTER_ID}/namespaces/')
        response_data = response.json()
        namespaces = getitems(response_data, 'data.manifest.items')
        assert len(namespaces) == 1
        assert getitems(namespaces[0], 'metadata.name') == COMMON_CLUSTER_NS_NAME
        client.delete_ignore_nonexistent(name=COMMON_CLUSTER_NS_NAME)

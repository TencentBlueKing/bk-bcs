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
import pytest

from backend.tests.conftest import TEST_PROJECT_ID

pytestmark = pytest.mark.django_db
API_URL_PREFIX = f'/api/projects/{TEST_PROJECT_ID}/cc'


def fake_fetch_topo(*args, **kwargs):
    """返回测试用 topo 数据"""
    return [
        {
            "default": 0,
            "bk_obj_name": "业务",
            "bk_obj_id": "biz",
            "child": [
                {
                    "default": 0,
                    "bk_obj_name": "集群",
                    "bk_obj_id": "set",
                    "child": [
                        {
                            "default": 0,
                            "bk_obj_name": "模块",
                            "bk_obj_id": "module",
                            "child": [],
                            "bk_inst_id": 1001,
                            "bk_inst_name": "空闲机",
                        }
                    ],
                    "bk_inst_id": 1,
                    "bk_inst_name": "BCS-K8S-1000",
                },
                {
                    "default": 0,
                    "bk_obj_name": "集群",
                    "bk_obj_id": "set",
                    "child": [
                        {
                            "default": 0,
                            "bk_obj_name": "模块",
                            "bk_obj_id": "module",
                            "child": [],
                            "bk_inst_id": 5003,
                            "bk_inst_name": "bcs-master",
                        },
                        {
                            "default": 0,
                            "bk_obj_name": "模块",
                            "bk_obj_id": "module",
                            "child": [],
                            "bk_inst_id": 5002,
                            "bk_inst_name": "bcs-node",
                        },
                    ],
                    "bk_inst_id": 5001,
                    "bk_inst_name": "BCS-K8S-1001",
                },
            ],
            "bk_inst_id": 10001,
            "bk_inst_name": "BCS",
        }
    ]


def fake_fetch_all_hosts(*args, **kwargs):
    """返回测试用主机数据（省略非必须字段）"""
    return [
        # 被使用的
        {
            "bk_cloud_id": 0,
            "bk_host_id": 1,
            "bk_host_innerip": "127.0.0.1",
            "svr_device_class": "S1234",
        },
        {
            "bk_cloud_id": 0,
            "bk_host_id": 2,
            "bk_host_innerip": "127.0.0.16",
            "svr_device_class": "S1234",
        },
        # agent 异常的
        {
            "bk_cloud_id": 0,
            "bk_host_id": 3,
            "bk_host_innerip": "127.0.0.2,127.0.0.3",
            "svr_device_class": "S1234ABC",
        },
        # Docker 机，不可用
        {
            "bk_cloud_id": 0,
            "bk_host_id": 4,
            "bk_host_innerip": "127.0.0.4,127.0.0.5",
            "svr_device_class": "D700249",
        },
        # 可以使用的
        {
            "bk_cloud_id": 0,
            "bk_host_id": 5,
            "bk_host_innerip": "127.0.0.6",
            "svr_device_class": "S1234",
        },
    ]


def fake_get_project_cluster_resource(*args, **kwargs):
    """返回测试用的项目，集群数据"""
    return [
        {
            "cluster_list": [
                {
                    "id": "BCS-K8S-1001",
                    "is_public": False,
                    "name": "测试用集群",
                    "namespace_list": [{"id": 101, "name": "default"}],
                },
            ],
            "code": "service_test",
            "id": "b3776666666666666667037f",
            "name": "容器服务测试",
        }
    ]


def fake_get_all_cluster_hosts(*args, **kwargs):
    """返回测试用的集群节点信息"""
    return [
        {"cluster_id": "BCS-K8S-1001", "inner_ip": "127.0.0.1", "status": "normal"},
        {"cluster_id": "BCS-K8S-1001", "inner_ip": "127.0.0.16", "status": "normal"},
    ]


def fake_get_agent_status(*args, **kwargs):
    """返回测试用 Agent 状态信息"""
    return [
        {"ip": "127.0.0.1", "bk_cloud_id": 0, "bk_agent_alive": 1},
        {"ip": "127.0.0.2", "bk_cloud_id": 0, "bk_agent_alive": 0},
        {"ip": "127.0.0.4", "bk_cloud_id": 0, "bk_agent_alive": 1},
        {"ip": "127.0.0.6", "bk_cloud_id": 0, "bk_agent_alive": 1},
        {"ip": "127.0.0.16", "bk_cloud_id": 0, "bk_agent_alive": 1},
    ]


class TestCCAPI:
    """测试 CMDB API 相关接口"""

    @mock.patch('backend.container_service.clusters.cc_host.views.cc.BizTopoQueryService.fetch', new=fake_fetch_topo)
    def test_get_biz_inst_topology(self, api_client):
        """测试创建资源接口"""
        response = api_client.get(f'{API_URL_PREFIX}/topology/')
        assert response.json()['code'] == 0

    @pytest.fixture()
    def patch_list_hosts_api_call(self):
        """mock cmdb, paas_cc, gse 接口"""
        with mock.patch(
            'backend.container_service.clusters.cc_host.views.cc.get_application_name',
            new=lambda *args, **kwargs: 'test-app-name',
        ), mock.patch(
            'backend.container_service.clusters.cc_host.views.cc.HostQueryService.fetch_all',
            new=fake_fetch_all_hosts,
        ), mock.patch(
            'backend.container_service.clusters.cc_host.utils.paas_cc.get_project_cluster_resource',
            new=fake_get_project_cluster_resource,
        ), mock.patch(
            'backend.container_service.clusters.cc_host.utils.paas_cc.get_all_cluster_hosts',
            new=fake_get_all_cluster_hosts,
        ), mock.patch(
            'backend.container_service.clusters.cc_host.utils.gse.get_agent_status', new=fake_get_agent_status
        ):
            yield

    def test_list_hosts(self, api_client, patch_list_hosts_api_call):
        """测试获取资源列表接口"""
        params = {'limit': 4, 'offset': 0, 'ip_list': [], 'set_id': 5001, 'module_id': 5003}
        response = api_client.post(f'{API_URL_PREFIX}/hosts/', data=params)
        assert response.json()['code'] == 0
        resp_data = response.json()['data']
        assert resp_data['count'] == 5
        assert len(resp_data['results']) == 4
        assert resp_data['results'][-1]['is_used']
        assert resp_data['results'][-1]['cluster_id'] == 'BCS-K8S-1001'

    def test_list_hosts_with_fuzzy_ip_match(self, api_client, patch_list_hosts_api_call):
        """测试按 ip 过滤（模糊）"""
        # 匹配 127.0.0.1， 127.0.0.16
        params = {'limit': 10, 'offset': 0, 'ip_list': ['127.0.0.1'], 'fuzzy': True}
        response = api_client.post(f'{API_URL_PREFIX}/hosts/', data=params)
        assert response.json()['code'] == 0
        resp_data = response.json()['data']
        assert resp_data['count'] == 2
        assert set([h['bk_host_innerip'] for h in resp_data['results']]) == {'127.0.0.1', '127.0.0.16'}

    def test_list_hosts_with_ip_match(self, api_client, patch_list_hosts_api_call):
        """测试按 ip 过滤（精确）"""
        params = {'limit': 10, 'offset': 0, 'ip_list': ['127.0.0.1', '127.0.0.2']}
        response = api_client.post(f'{API_URL_PREFIX}/hosts/', data=params)
        assert response.json()['code'] == 0
        resp_data = response.json()['data']
        assert resp_data['count'] == 2
        assert set([h['bk_host_innerip'] for h in resp_data['results']]) == {'127.0.0.1', '127.0.0.2,127.0.0.3'}

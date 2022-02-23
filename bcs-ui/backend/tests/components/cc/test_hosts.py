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

from backend.components.cc import BizTopoQueryService, HostQueryService, get_has_perm_hosts

# 测试用 search_biz_inst_topo 返回结果
FAKE_TOPO_RESP = [
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
                        "bk_inst_id": 5003,
                        "bk_inst_name": "bcs-master",
                    }
                ],
                "bk_inst_id": 5001,
                "bk_inst_name": "BCS-K8S-1001",
            }
        ],
        "bk_inst_id": 10001,
        "bk_inst_name": "BCS",
    }
]

# 测试用 get_biz_internal_module 返回结果
FAKE_INTERNAL_MODULE_RESP = {
    "bk_set_id": 1,
    "bk_set_name": "空闲机池",
    "module": [
        {
            "bk_module_id": 11,
            "bk_module_name": "空闲机",
        },
        {
            "bk_module_id": 12,
            "bk_module_name": "故障机",
        },
    ],
}

# 测试用 list_biz_hosts 返回结果
FAKE_HOSTS_RESP = {
    # 设置为 501，刚好触发查询两次逻辑
    "count": 501,
    "info": [
        {
            "bk_cloud_id": 0,
            "bk_host_id": 1,
            "bk_host_innerip": "127.0.0.1",
            "svr_device_class": "S1234",
            "operator": "admin",
        },
        {
            "bk_cloud_id": 0,
            "bk_host_id": 2,
            "bk_host_innerip": "127.0.0.16",
            "svr_device_class": "S1234",
        },
    ],
}


class TestComponentCCHosts:
    @pytest.fixture(autouse=True)
    def patch_api_call(self):
        with mock.patch(
            'backend.components.cc.hosts.BkCCClient.search_biz_inst_topo', return_value=FAKE_TOPO_RESP
        ), mock.patch(
            'backend.components.cc.hosts.BkCCClient.list_biz_hosts', return_value=FAKE_HOSTS_RESP
        ), mock.patch(
            'backend.components.cc.hosts.BkCCClient.get_biz_internal_module', return_value=FAKE_INTERNAL_MODULE_RESP
        ), mock.patch(
            'backend.components.cc.hosts.get_app_maintainers', return_value=[]
        ):
            yield

    def test_get_has_perm_hosts(self):
        # 通过 mock get_app_maintainers 走 _get_hosts_by_operator 逻辑
        ret = get_has_perm_hosts(1001, 'admin')
        assert len(ret) == 2

    def test_search_biz_inst_topo(self):
        ret = BizTopoQueryService('admin', 1001).fetch()
        assert ret[0]['child'][0]['child'][0]['bk_inst_id'] == 11
        assert ret[0]['child'][1]['bk_inst_id'] == 5001

    def test_fetch_all_hosts(self):
        ret = HostQueryService('admin', 1001).fetch_all()
        assert len(ret) == 4

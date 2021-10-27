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

from backend.components.cc.client import BkCCClient, PageData

# 测试用 search_business 返回结果
FAKE_BIZS_RESP = {
    "code": 0,
    "data": {
        "count": 1,
        "info": [
            {
                "bs2_name_id": 1,
                "default": 0,
            }
        ],
    },
}

# 测试用 search_biz_inst_topo 返回结果
FAKE_TOPO_RESP = {
    "code": 0,
    "data": [
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
    ],
}

# 测试用 get_biz_internal_module 放回结果
FAKE_INTERNAL_MODULE_RESP = {
    "code": 0,
    "data": {
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
    },
}

# 测试用 list_biz_hosts 返回结果
FAKE_HOSTS_RESP = {
    "code": 0,
    "data": {
        "count": 2,
        "info": [
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
        ],
    },
}


class TestBkCCClient:
    def test_search_business(self, request_user, requests_mock):
        requests_mock.post(ANY, json=FAKE_BIZS_RESP)
        client = BkCCClient(request_user.username)
        data = client.search_business(PageData(), ["bs2_name_id"], {"bk_biz_id": 1})
        assert data["info"][0]["bs2_name_id"] == 1
        assert requests_mock.called

    def test_search_biz_inst_topo(self, request_user, requests_mock):
        requests_mock.post(ANY, json=FAKE_TOPO_RESP)
        topo = BkCCClient(request_user.username).search_biz_inst_topo(1)
        assert topo[0]['child'][0]['bk_inst_id'] == 5001
        assert requests_mock.called

    def test_get_biz_internal_module(self, request_user, requests_mock):
        requests_mock.post(ANY, json=FAKE_INTERNAL_MODULE_RESP)
        internal_module = BkCCClient(request_user.username).get_biz_internal_module(1)
        assert internal_module['bk_set_name'] == '空闲机池'
        assert internal_module['module'][0]['bk_module_id'] == 11
        assert requests_mock.called

    def test_list_biz_hosts(self, request_user, requests_mock):
        requests_mock.post(ANY, json=FAKE_HOSTS_RESP)
        resp = BkCCClient(request_user.username).list_biz_hosts(
            1001, PageData(), bk_set_ids=[], bk_module_ids=[], fields=[]
        )
        assert resp['count'] == 2
        assert resp['info'][0]['bk_host_innerip'] == '127.0.0.1'
        assert requests_mock.called

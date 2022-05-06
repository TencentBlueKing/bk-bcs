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

from backend.components.cc import (
    AppQueryService,
    fetch_has_maintain_perm_apps,
    get_app_maintainers,
    get_application_name,
)

FAKE_BIZS_INFO = {
    # 设置会 201，刚好可以触发全量查询时候查询两次
    "count": 201,
    "info": [
        {
            "bs2_name_id": 1,
            "bk_biz_id": 1001,
            "bk_biz_name": "测试业务",
            "default": 0,
            "bk_biz_maintainer": "admin,admin1",
        }
    ],
}


class TestComponentCCBusiness:
    @pytest.fixture(autouse=True)
    def patch_api_call(self):
        with mock.patch('backend.components.cc.business.BkCCClient.search_business', return_value=FAKE_BIZS_INFO):
            yield

    def test_fetch_has_maintain_perm_apps(self):
        ret = fetch_has_maintain_perm_apps('admin')
        assert ret == [{'id': 1001, "name": "测试业务"}, {'id': 1001, "name": "测试业务"}]

    def test_get_single_app_info(self):
        """测试 AppQueryService.get"""
        ret = AppQueryService('admin').get(1001)
        assert ret == FAKE_BIZS_INFO['info'][0]

    def test_get_application_name(self):
        ret = get_application_name(1001)
        assert ret == '测试业务'

    def test_get_app_maintainers(self):
        ret = get_app_maintainers('admin', 1001)
        assert ret == ['admin', 'admin1']

    def test_fetch_all_apps(self):
        """测试 AppQueryService.fetch_all"""
        ret = AppQueryService('admin').fetch_all()
        assert len(ret) == 2

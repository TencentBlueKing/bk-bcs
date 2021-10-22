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
from unittest.mock import patch

from backend.container_service.projects.cmdb import is_biz_maintainer


@patch("backend.container_service.projects.cmdb.cc.get_app_maintainers", return_value=["admin", "admin1"])
def test_is_biz_maintainers(mock_get_app_maintainers):
    assert not is_biz_maintainer(1, "demo1")
    assert is_biz_maintainer(1, "admin1")

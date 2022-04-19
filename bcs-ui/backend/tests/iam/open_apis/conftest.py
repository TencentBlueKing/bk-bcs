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


@pytest.fixture(autouse=True)
def patch4resource_api():
    with mock.patch(
        'backend.iam.open_apis.authentication.IamBasicAuthentication.authenticate', new=lambda *args, **kwargs: None
    ), mock.patch(
        'backend.iam.open_apis.providers.utils.get_client_access_token',
        new=lambda *args, **kwargs: {"access_token": "test"},
    ), mock.patch(
        'backend.iam.open_apis.providers.namespace.get_shared_clusters', new=lambda *args, **kwargs: []
    ):
        yield

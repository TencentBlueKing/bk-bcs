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
from mock import patch

from backend.helm.releases.utils.release_secret import get_release_detail


def test_get_release_detail(ctx_cluster, default_namespace, release_name, revision, parsed_release_data):
    with patch(
        "backend.helm.releases.utils.release_secret.list_namespaced_releases",
        return_value=[parsed_release_data],
    ):
        release_detail = get_release_detail(ctx_cluster, default_namespace, release_name)
    assert release_detail["name"] == release_name
    assert release_detail["namespace"] == default_namespace
    assert release_detail["version"] == revision

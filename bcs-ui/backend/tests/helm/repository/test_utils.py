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

from backend.helm.repository.constants import RepoCategory
from backend.helm.repository.utils import is_imported_repo

FAKE_PROJECT_CODE = "test"
FAKE_REPO_NAME = "test"


@pytest.mark.parametrize(
    "repos,expected",
    [
        ([{"name": "test", "category": RepoCategory.LOCAL}], False),
        ([{"name": "test", "category": RepoCategory.REMOTE}], True),
        (
            [
                {
                    "name": "test",
                    "category": RepoCategory.COMPOSITE,
                    "configuration": {"proxy": {"channelList": [{"name": FAKE_REPO_NAME}]}},
                }
            ],
            True,
        ),
    ],
)
def test_is_imported_repo(request_user, repos, expected):
    with mock.patch("backend.helm.repository.utils.BkRepoClient.list_project_repos", return_value=repos):
        assert (
            is_imported_repo(request_user.token.access_token, request_user.username, FAKE_PROJECT_CODE, FAKE_REPO_NAME)
            == expected
        )

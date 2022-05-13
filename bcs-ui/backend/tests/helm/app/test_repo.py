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

import pytest

from backend.helm.app import repo
from backend.tests.bcs_mocks.fake_bk_repo import FakeBkRepoMod
from backend.utils import FancyDict

pytestmark = pytest.mark.django_db

FAKE_USER = FancyDict(username="admin", token=FancyDict(access_token="access_token"))
FAKE_PROJECT = FancyDict(
    project_id="project_id",
    project_code="project_code",
    project_name="project_name",
    english_name="project_code",
    description="this is a test",
)


@patch("backend.helm.app.repo.bk_repo.BkRepoClient", new=FakeBkRepoMod)
def test_get_or_create_private_repo(bk_user):
    detail = repo.get_or_create_private_repo(bk_user, FAKE_PROJECT)
    assert detail.name == FAKE_PROJECT.project_code
    assert detail.project_id == FAKE_PROJECT.project_id
    assert repo.Repository.objects.filter(name=FAKE_PROJECT.project_code, project_id=FAKE_PROJECT.project_id).exists

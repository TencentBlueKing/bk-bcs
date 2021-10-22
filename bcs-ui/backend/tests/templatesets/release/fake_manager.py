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
from backend.templatesets.release.manager import ReleaseResourceManager


def fake_async_run(tasks):
    for t in tasks:
        t()


class FakeApi:
    def update_or_create(self, *args, **kwargs):
        pass

    def delete_ignore_nonexistent(self, *args, **kwargs):
        pass


class FakeReleaseResourceManager(ReleaseResourceManager):
    def _get_api(self, kind, api_version):
        return FakeApi()

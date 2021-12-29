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
from iam import Request

from backend.iam.permissions.perm import Permission

from ..permissions import roles


class FakeNamespaceIAM:
    def is_allowed(self, request: Request) -> bool:
        if request.subject.id in [roles.ADMIN_USER, roles.NAMESPACE_NO_CLUSTER_PROJECT_USER]:
            return True
        return False

    def is_allowed_with_cache(self, request: Request) -> bool:
        return self.is_allowed(request)


class FakeNamespacePermission(Permission):
    iam = FakeNamespaceIAM()

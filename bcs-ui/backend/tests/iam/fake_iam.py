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
from typing import Dict, List

from iam import Request

from backend.iam.permissions.perm import Permission
from backend.iam.permissions.request import ResourceRequest
from backend.iam.permissions.resources.project import ProjectAction

from .permissions import roles


class FakeProjectIAM:
    def is_allowed(self, request: Request) -> bool:
        if request.subject.id in [roles.ADMIN_USER, roles.PROJECT_CLUSTER_USER, roles.PROJECT_NO_CLUSTER_USER]:
            return True

        if request.subject.id == roles.PROJECT_NO_VIEW_USER:
            if request.action.id == ProjectAction.VIEW:
                return False
            return True

        return False

    def is_allowed_with_cache(self, request: Request) -> bool:
        return self.is_allowed(request)


class FakeProjectPermission(Permission):
    iam = FakeProjectIAM()


class FakeClusterIAM:
    def is_allowed(self, request: Request) -> bool:
        if request.subject.id in [
            roles.ADMIN_USER,
            roles.CLUSTER_USER,
            roles.PROJECT_CLUSTER_USER,
            roles.CLUSTER_NO_PROJECT_USER,
        ]:
            return True
        return False

    def is_allowed_with_cache(self, request: Request) -> bool:
        return self.is_allowed(request)


class FakeClusterPermission(Permission):
    iam = FakeClusterIAM()


class FakeNamespaceIAM:
    def is_allowed(self, request: Request) -> bool:
        if request.subject.id in [roles.ADMIN_USER, roles.NAMESPACE_NO_CLUSTER_PROJECT_USER]:
            return True
        return False

    def is_allowed_with_cache(self, request: Request) -> bool:
        return self.is_allowed(request)


class FakeNamespacePermission(Permission):
    iam = FakeNamespaceIAM()


class FakeTemplatesetIAM:
    def __init__(self, *args, **kwargs):
        """"""

    def is_allowed(self, request: Request) -> bool:
        if request.subject.id in [
            roles.ADMIN_USER,
            roles.TEMPLATESET_USER,
            roles.PROJECT_TEMPLATESET_USER,
            roles.TEMPLATESET_NO_PROJECT_USER,
        ]:
            return True
        return False

    def is_allowed_with_cache(self, request: Request) -> bool:
        return self.is_allowed(request)


class FakeTemplatesetPermission(Permission):
    iam = FakeTemplatesetIAM()


class FakeIAMClient:
    def resource_type_allowed(self, username: str, action_id: str, use_cache: bool = False) -> bool:
        return action_id == 'project_create'

    def resource_inst_allowed(
        self, username: str, action_id: str, res_request: ResourceRequest, use_cache: bool = False
    ) -> bool:
        return action_id in ['cluster_create', 'cluster_view']

    def resource_type_multi_actions_allowed(self, username: str, action_ids: List[str]) -> Dict[str, bool]:
        return {action_id: self.resource_type_allowed(username, action_id) for action_id in action_ids}

    def resource_inst_multi_actions_allowed(
        self, username: str, action_ids: List[str], res_request: ResourceRequest
    ) -> Dict[str, bool]:
        multi = {}
        for action in action_ids:
            is_allowed = False
            if action in ['cluster_create', 'cluster_view']:
                is_allowed = True
            multi[action] = is_allowed
        return multi

    def batch_resource_multi_actions_allowed(
        self, username: str, action_ids: List[str], res_request: ResourceRequest
    ) -> Dict[str, Dict[str, bool]]:
        res = res_request.res
        if isinstance(res, str):
            res = [res]

        perms = {}

        for idx, r_id in enumerate(res):
            if idx % 2 == 0:
                p = {action_id: False for action_id in action_ids}
            else:
                p = {action_id: True for action_id in action_ids}
            perms[r_id] = p

        return perms

    def grant_resource_creator_actions(self, username: str, data: Dict) -> (bool, str):
        return True, 'success'

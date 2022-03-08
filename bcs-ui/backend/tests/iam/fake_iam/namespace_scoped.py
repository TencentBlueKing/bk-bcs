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

from iam import Resource

from backend.iam.permissions.perm import Permission
from backend.iam.permissions.resources.namespace_scoped import NamespaceScopedAction

from ..permissions import roles


class FakeNamespaceScopedPermission(Permission):
    def resource_inst_multi_actions_allowed(
        self, username: str, action_ids: List[str], resources: List[Resource]
    ) -> Dict[str, bool]:
        if username == roles.ADMIN_USER:
            return {action_id: True for action_id in action_ids}
        if username == roles.NAMESPACE_SCOPED_NO_VIEW_USER:
            multi = {action_id: True for action_id in action_ids}
            multi[NamespaceScopedAction.VIEW] = False
            return multi
        return {action_id: False for action_id in action_ids}

    def batch_resource_multi_actions_allowed(
        self, username: str, action_ids: List[str], resources: List[Resource]
    ) -> Dict[str, Dict[str, bool]]:
        if username == roles.ADMIN_USER:
            actions_allowed = {action_id: True for action_id in action_ids}
        elif username == roles.NAMESPACE_SCOPED_NO_VIEW_USER:
            multi = {action_id: True for action_id in action_ids}
            multi[NamespaceScopedAction.VIEW] = False
            actions_allowed = multi
        else:
            actions_allowed = {action_id: False for action_id in action_ids}

        return {res.id: actions_allowed for res in resources}

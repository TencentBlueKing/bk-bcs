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
import logging

from backend.iam.permissions.resources.cluster import ClusterPermCtx, ClusterPermission

logger = logging.getLogger(__name__)


def with_function_controller_check(permission_cls):
    def wrapper(cls):
        cls.permission_classes = (*cls.permission_classes, permission_cls)
        return cls

    return wrapper


def check_cluster_perm(user, project_id, cluster_id, raise_exception=True, request=None):
    perm_ctx = ClusterPermCtx(username=user.username, project_id=project_id, cluster_id=cluster_id)
    ClusterPermission().can_view(perm_ctx)

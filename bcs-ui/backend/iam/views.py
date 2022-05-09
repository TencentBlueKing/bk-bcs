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

from rest_framework import permissions
from rest_framework.exceptions import ValidationError
from rest_framework.response import Response

from backend.bcs_web import viewsets

from .perm_maker import make_perm_ctx, make_res_permission
from .permissions.client import IAMClient
from .permissions.exceptions import AttrValidationError, PermissionDeniedError
from .request_maker import make_request_resources
from .serializers import ResourceActionSLZ, ResourceMultiActionsSLZ

logger = logging.getLogger(__name__)


class UserPermsViewSet(viewsets.SystemViewSet):
    permission_classes = (permissions.IsAuthenticated,)

    def get_perms(self, request):
        """查询多个 action_id 的权限"""
        validated_data = self.params_validate(ResourceMultiActionsSLZ)
        perm_ctx = validated_data['perm_ctx']

        client = IAMClient()

        # 资源实例无关
        if not perm_ctx:
            perms = client.resource_type_multi_actions_allowed(request.user.username, validated_data['action_ids'])
            return Response({'perms': perms})

        # 资源实例相关
        resource_type = perm_ctx.pop('resource_type')
        try:
            request_resources = make_request_resources(resource_type, **perm_ctx)
        except AttrValidationError as e:
            raise ValidationError(e)

        perms = client.resource_inst_multi_actions_allowed(
            request.user.username, validated_data['action_ids'], request_resources
        )
        return Response({'perms': perms})

    def get_perm_by_action_id(self, request, action_id):
        """查询指定 action_id 的权限"""
        validated_data = self.params_validate(ResourceActionSLZ, action_id=action_id)

        try:
            perm_ctx = make_perm_ctx(action_id, request.user.username, **validated_data['perm_ctx'])
        except AttrValidationError as e:
            raise ValidationError(e)

        permission = make_res_permission(action_id)
        try:
            # 调用 permission.can_xx 方法
            # rsplit 用于从 action_id 中提取动词, 如 namespace_scoped_view 中提取出 view
            verb = action_id.rsplit('_', 1)[-1]
            getattr(permission, f'can_{verb}')(perm_ctx)
        except AttributeError:
            raise ValidationError(f'action_id({action_id}) not supported')
        except AttrValidationError as e:
            raise ValidationError(e)
        except PermissionDeniedError as e:
            return Response({'perms': {action_id: False, 'apply_url': e.data['perms']['apply_url']}})

        return Response({'perms': {action_id: True}})

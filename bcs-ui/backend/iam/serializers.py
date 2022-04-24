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
from rest_framework import serializers
from rest_framework.exceptions import ValidationError

from .permissions.resources.constants import ResourceType


class ResourceActionSLZ(serializers.Serializer):
    action_id = serializers.CharField()
    perm_ctx = serializers.JSONField(default=dict)

    def validate(self, data):
        # data['action_id'] like namespace_view, namespace_scoped_view
        resource_type = data['action_id'].split('_', 1)[0]
        if resource_type not in ResourceType.get_values():
            raise ValidationError(f"invalid action_id({data['action_id']})")
        return data


class ResourceMultiActionsSLZ(serializers.Serializer):
    action_ids = serializers.ListField(child=serializers.CharField(), min_length=1)
    perm_ctx = serializers.JSONField(default=dict)

    def validate_perm_ctx(self, perm_ctx):
        if not perm_ctx:
            return perm_ctx

        resource_type = perm_ctx.get('resource_type')
        if resource_type not in ResourceType.get_values():
            raise ValidationError(f"invalid resource_type {resource_type}")

        return perm_ctx

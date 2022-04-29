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
from urllib.parse import urlparse

from django.utils.translation import ugettext_lazy as _
from rest_framework import serializers
from rest_framework.exceptions import PermissionDenied, ValidationError

from .models import Token, make_random_key


class TokenSLZ(serializers.ModelSerializer):
    """新增/查询/删除 三种操作"""

    user = serializers.HiddenField(default=serializers.CurrentUserDefault(), write_only=True)
    username = serializers.CharField(read_only=True)
    config = serializers.JSONField(
        initial={},
        write_only=True,
        label="Config",
        help_text="JSON format data",
        style={"base_template": "textarea.html", "rows": 10},
    )

    def create(self, validated_data):
        return self.Meta.model.objects.make_token(**validated_data)

    class Meta:
        model = Token
        fields = (
            "user",
            "username",
            "config",
            "key",
            "name",
            "kind",
            "config",
            "description",
            "id",
            "maintainers",
        )

        read_only_fields = (
            "key",
            "username",
        )


class TokenUpdateSLZ(serializers.ModelSerializer):
    """仅用于更新 token"""

    user = serializers.HiddenField(default=serializers.CurrentUserDefault(), write_only=True)

    def update(self, instance, validated_data):
        """更新key"""
        if instance.key != validated_data["key"]:
            raise PermissionDenied(_("输入的key与当前key不匹配"))

        instance.key = make_random_key()
        instance.save()
        return instance

    class Meta:
        model = Token
        fields = (
            "user",
            "username",
            "config",
            "key",
            "name",
            "kind",
            "config",
            "description",
            "id",
            "maintainers",
        )

        read_only_fields = (
            "username",
            "config",
            "name",
            "kind",
            "config",
            "description",
            "id",
            "maintainers",
        )

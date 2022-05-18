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

from django.utils.translation import ugettext_lazy as _
from rest_framework import serializers
from rest_framework.exceptions import ValidationError


class RepoParamsSLZ(serializers.Serializer):
    """仓库参数"""

    name = serializers.CharField()
    is_public = serializers.BooleanField(help_text="是否为公有源")
    url = serializers.CharField(help_text="访问仓库的地址")
    username = serializers.CharField(required=False, help_text="私有源时需要用户名")
    password = serializers.CharField(required=False, help_text="私有源时需要密码")

    def validate(self, data):
        if data["is_public"] and not (data["username"] and data["password"]):
            raise ValidationError(_("参数【is_public】为真时，参数【username】和【password】不能为空"))

        return data

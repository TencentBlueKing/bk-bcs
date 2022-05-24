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

from backend.components.bk_repo import RepoConfig
from backend.helm.repository.constants import RepoCategory, RepoType


class RepoParamsSLZ(serializers.Serializer):
    """仓库参数"""

    name = serializers.CharField()
    description = serializers.CharField(default="", allow_blank=True)
    is_public = serializers.BooleanField(default=False, help_text="是否为公有源")
    url = serializers.CharField(help_text="访问仓库的地址")
    username = serializers.CharField(default="", help_text="私有源时需要用户名", allow_blank=True)
    password = serializers.CharField(default="", help_text="私有源时需要密码", allow_blank=True)

    def validate(self, data):
        if data["is_public"] and not (data["username"] and data["password"]):
            raise ValidationError(_("参数【is_public】为真时，参数【username】和【password】不能为空"))

        # 添加额外参数，便于后续组装请求参数
        data.update(
            {
                "type": RepoType.HELM,
                "category": RepoCategory.COMPOSITE,
                "public": False,  # 纳管的仓库不允许 public
                "configuration": RepoConfig(
                    type=RepoCategory.COMPOSITE.lower(),
                    proxy={
                        "channelList": [
                            {
                                "public": data["is_public"],
                                "name": data["name"],
                                "url": data["url"],
                                "username": data["username"],
                                "password": data["password"],
                            }
                        ]
                    },
                ),
            }
        )

        return data

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
import rest_framework.authtoken.models
from django.conf import settings
from django.db import models
from django.utils.crypto import get_random_string
from django.utils.translation import ugettext_lazy as _
from jsonfield import JSONField
from six import python_2_unicode_compatible

from .managers import TokenManager
from .providers import provider_choice


def make_random_key():
    return get_random_string(63)


@python_2_unicode_compatible
class Token(models.Model):
    # key is no longer primary key, but still indexed and unique
    key = models.CharField(_("Key"), max_length=64, db_index=True, unique=True, default=make_random_key)
    created = models.DateTimeField(_("Created"), auto_now_add=True)
    # since real user model is abstract, we use username field here
    username = models.CharField(max_length=64, db_index=True)

    # maintainers 仅用于平台方联系使用方时使用(比如：认证/鉴权体系需要升级)
    maintainers = models.CharField(_("Maintainers"), max_length=256, help_text="multiple items split with ;")
    name = models.CharField(_("Name"), max_length=64, help_text="ex: {project_id}-{cluster_id}-{app_id}-helm-app")
    description = models.CharField(_("Description"), max_length=256)

    # kind 用于指定应用场景, 不同kind用于处理不同应用场景的鉴权信息
    kind = models.CharField(_("Kind"), max_length=32, choices=provider_choice)
    # config 配合kind完成鉴权，比如 kind=helm-app-update 类型时，config 会存放 helm app id
    config = JSONField(null=True, default={})

    objects = TokenManager()

    class Meta:
        unique_together = (('username', 'name'),)

    def validate_request_data(self, request_data):
        Token.objects.validate_request_data(token=self, request_data=request_data)

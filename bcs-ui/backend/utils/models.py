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

common utils
"""
from django.db import models
from django.utils import timezone
from django.utils.translation import ugettext_lazy as _


class BaseTSModel(models.Model):
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)

    class Meta:
        abstract = True


class BaseModel(models.Model):
    """Model with 'created' and 'updated' fields."""

    # 根据用户管理的限制，允许creator和updater最大长度为64
    creator = models.CharField("创建者", max_length=64)
    updator = models.CharField("修改者", max_length=64)
    created = models.DateTimeField(auto_now_add=True)
    updated = models.DateTimeField(auto_now=True)
    is_deleted = models.BooleanField(default=False)
    deleted_time = models.DateTimeField(null=True, blank=True)

    @property
    def created_display(self):
        # 转换成本地时间
        t = timezone.localtime(self.created)
        return t.strftime("%Y-%m-%d %H:%M:%S")

    @property
    def updated_display(self):
        # 转换成本地时间
        t = timezone.localtime(self.updated)
        return t.strftime("%Y-%m-%d %H:%M:%S")

    @property
    def updated_display_short(self):
        # 转换成本地时间
        t = timezone.localtime(self.updated)
        return t.strftime("%Y-%m-%d %H:%M")

    class Meta:
        abstract = True

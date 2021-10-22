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
from django.db import models
from django.utils.translation import ugettext_lazy as _

from backend.utils.models import BaseModel


class TlsCertManager(models.Manager):
    """Manager for UserGroup"""

    def get_queryset(self):
        return super().get_queryset().filter(is_deleted=False)


class TlsCert(BaseModel):
    """
    tls 证书
    """

    project_id = models.CharField(_("项目ID"), max_length=32)
    # 证书名称不能为空，只支持英文大小写、数字、下划线和英文句号
    name = models.CharField(_("名称"), max_length=128)
    cert = models.TextField(_("证书内容"), blank=True, null=True)
    key = models.TextField(_("证书内容"), blank=True, null=True)

    objects = TlsCertManager()
    default_objects = models.Manager()

    class Meta:
        unique_together = ("project_id", "name")
        ordering = ('-id',)

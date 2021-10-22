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


class TemplateManager(models.Manager):
    """Manager for Template"""

    def get_queryset(self):
        return super().get_queryset().filter(is_deleted=False)


class ShowVersionManager(models.Manager):
    """Manager for ShowVersionManager"""

    def get_queryset(self):
        return super().get_queryset().filter(is_deleted=False)

    def get_latest_by_template(self, template_id):
        return self.filter(template_id=template_id).order_by('-updated').first()


class VersionedEntityManager(models.Manager):
    def get_latest_by_template(self, template_id):
        return self.filter(template_id=template_id).order_by('-updated').first()

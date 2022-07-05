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

from backend.utils.models import BaseModel

from .constants import LogSourceType


class LogIndexSet(models.Model):
    project_id = models.CharField(max_length=32, unique=True)
    bk_biz_id = models.IntegerField()
    std_index_set_id = models.IntegerField()
    file_index_set_id = models.IntegerField()

    @classmethod
    def safe_create(cls, project_id: str, bk_biz_id: int, **kwargs) -> 'LogIndexSet':
        defaults = {}

        std_index_set_id = kwargs.get('std_index_set_id', 0)
        if std_index_set_id:
            defaults['std_index_set_id'] = std_index_set_id

        file_index_set_id = kwargs.get('file_index_set_id', 0)
        if file_index_set_id:
            defaults['file_index_set_id'] = file_index_set_id

        obj, _ = cls.objects.update_or_create(project_id=project_id, bk_biz_id=bk_biz_id, defaults=defaults)

        return obj


class LogCollectMetadata(BaseModel):
    project_id = models.CharField(max_length=32)
    cluster_id = models.CharField(max_length=32)
    namespace = models.CharField(max_length=63)
    log_source_type = models.CharField(
        choices=LogSourceType.get_choices(), max_length=32, default=LogSourceType.SELECTED_CONTAINERS.value
    )
    config_id = models.IntegerField(help_text='alias rule_id', unique=True)
    config_name = models.CharField(max_length=255)

    class Meta:
        ordering = ['-updated']

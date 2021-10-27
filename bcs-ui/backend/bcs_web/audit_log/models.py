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
import json

from django.db import models

from . import constants


class UserActivityLog(models.Model):
    project_id = models.CharField(help_text='项目id', max_length=32, db_index=True, null=False, blank=False)

    activity_time = models.DateTimeField(auto_now_add=True)
    activity_type = models.CharField(
        help_text='操作类型', choices=constants.ActivityType.get_django_choices(), max_length=32, default=''
    )
    activity_status = models.CharField(
        help_text='操作状态', choices=constants.ActivityStatus.get_django_choices(), max_length=32, default=''
    )

    resource = models.CharField(help_text='操作对象', null=True, blank=True, max_length=512)
    resource_id = models.CharField(help_text='操作对象id', null=True, blank=True, max_length=256)
    resource_type = models.CharField(
        help_text='操作对象类型',
        null=True,
        blank=True,
        max_length=32,
        choices=constants.ResourceType.get_django_choices(),
    )

    user = models.CharField(help_text='发起者', max_length=64)
    description = models.TextField(help_text='描述', null=True, blank=True)
    extra = models.TextField(help_text='扩展')

    def save(self, *args, **kwargs):
        if isinstance(self.extra, dict):
            self.extra = json.dumps(self.extra)
        super().save(*args, **kwargs)


class UserActivityLogLabel(models.Model):
    activity_log = models.ForeignKey(UserActivityLog, on_delete=models.CASCADE)
    type = models.CharField(max_length=32, db_index=True)
    key = models.CharField(max_length=32, db_index=True)
    value = models.TextField()

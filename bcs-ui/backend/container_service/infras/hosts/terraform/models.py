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
from jsonfield import JSONField

from .constants import TaskStatus


class HostApplyTaskLog(models.Model):
    created = models.DateTimeField(auto_now_add=True)
    updated = models.DateTimeField(auto_now=True)
    project_id = models.CharField(max_length=32)
    task_id = models.CharField(max_length=64, null=True)
    task_url = models.CharField(max_length=256, null=True)
    operator = models.CharField(max_length=64, null=True)
    params = JSONField(null=True, default={})
    status = models.CharField(max_length=32, choices=TaskStatus.get_choices(), default=TaskStatus.RUNNING.value)
    is_finished = models.BooleanField(default=False)
    logs = JSONField(null=True, default={})

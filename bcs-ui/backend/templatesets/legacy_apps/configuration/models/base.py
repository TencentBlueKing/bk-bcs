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
import logging

from django.db import models
from django.utils import timezone

COPY_TEMPLATE = "-copy"
logger = logging.getLogger(__name__)

# 所有包含 pod 的资源
# TODO mark refactor 移到constants文件中
POD_RES_LIST = ['K8sDeployment', 'K8sDaemonSet', 'K8sJob', 'K8sStatefulSet']


def get_default_version():
    """版本号：默认为时间戳"""
    return timezone.localtime().strftime('%Y%m%d-%H%M%S')


class BaseModel(models.Model):
    """Model with 'created' and 'updated' fields."""

    # 根据用户管理的限制，允许creator和updater最大长度为64
    creator = models.CharField("创建者", max_length=64)
    updator = models.CharField("更新者", max_length=64)
    created = models.DateTimeField(auto_now_add=True)
    updated = models.DateTimeField(auto_now=True)
    is_deleted = models.BooleanField(default=False)
    deleted_time = models.DateTimeField(null=True, blank=True)

    class Meta:
        abstract = True

    def delete(self, *args, **kwargs):
        self.is_deleted = True
        self.deleted_time = timezone.now()
        self.save(update_fields=['is_deleted', 'deleted_time'])

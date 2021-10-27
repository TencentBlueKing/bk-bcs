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
from dataclasses import dataclass
from typing import List

from django.db import models
from jsonfield import JSONField

from backend.packages.blue_krill.data_types.enum import EnumField, StructuredEnum
from backend.templatesets.legacy_apps.configuration import models as config_models
from backend.utils.models import BaseModel


@dataclass
class ResourceData:
    kind: str
    name: str
    namespace: str
    manifest: dict
    version: str = ""
    revision: str = ""


@dataclass
class AppReleaseData:
    name: str
    project_id: str
    cluster_id: str
    namespace: str
    template_id: int
    resource_list: List[ResourceData]


class ReleaseStatus(str, StructuredEnum):
    # pending用于异步方案下的中间态
    PENDING = EnumField('pending', label='pending')
    DEPLOYED = EnumField('deployed', label='deployed')
    FAILED = EnumField('failed', label='failed')
    UNKNOWN = EnumField('unknown', label='unknown')


class AppRelease(BaseModel):
    name = models.CharField(max_length=256)
    project_id = models.CharField(max_length=32)
    cluster_id = models.CharField(max_length=32)
    namespace = models.CharField(max_length=64)
    status = models.CharField(choices=ReleaseStatus.get_choices(), default=ReleaseStatus.PENDING.value, max_length=32)
    message = models.TextField(default='')
    template_id = models.IntegerField("关联model Template")

    class Meta:
        db_table = 'templatesets_app_release'

    @property
    def template(self):
        return config_models.Template.objects.get(id=self.template_id)

    def update_status(self, status: str, message: str = 'success'):
        """
        更新release状态字段
        """
        self.status = status
        self.message = message
        self.save()


class ResourceInstance(BaseModel):
    app_release = models.ForeignKey(AppRelease, on_delete=models.SET_NULL, null=True)
    kind = models.CharField(max_length=64)
    name = models.CharField(max_length=255)
    namespace = models.CharField(max_length=64)
    manifest = JSONField()
    version = models.CharField("模板集版本名", max_length=255)
    revision = models.CharField("模板集版本名的修订版号", max_length=32)
    edited = models.BooleanField("是否在线编辑过", default=False)

    class Meta:
        db_table = 'templatesets_resource_instance'

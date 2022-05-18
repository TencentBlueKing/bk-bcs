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

from .constants import ToolStatus


class Tool(models.Model):
    """组件库中的可用组件. 由于通过 Helm Chart 管理, 因此存储了 Chart 信息"""

    chart_name = models.CharField(max_length=128, unique=True)
    name = models.CharField('组件名', max_length=64)
    default_version = models.CharField(max_length=64)
    default_values = models.TextField(null=True, blank=True, help_text='组件启用时需要额外设置的变量值，文本内容格式为 yaml')
    # 记录一些额外的启动命令如 --disable-openapi-validation 等
    extra_options = models.TextField(default='')
    namespace = models.CharField(max_length=64, default='bcs-system')
    description = models.TextField(help_text="组件功能介绍", null=True, blank=True)
    help_link = models.CharField(max_length=255, null=True, blank=True)
    logo = models.TextField('图片 logo', null=True, blank=True)


class InstalledTool(BaseModel):
    """记录已安装到集群中的组件信息"""

    tool = models.ForeignKey(Tool, on_delete=models.CASCADE, db_constraint=False)
    release_name = models.CharField(max_length=53)
    project_id = models.CharField(max_length=32)
    cluster_id = models.CharField(max_length=32)
    chart_url = models.CharField(max_length=255)
    values = models.TextField(null=True, blank=True, help_text="组件启用或更新时设置的变量值，文本内容格式为 yaml")
    namespace = models.CharField(max_length=64)
    status = models.CharField(choices=ToolStatus.get_choices(), default=ToolStatus.NOT_DEPLOYED, max_length=32)
    message = models.TextField('记录错误信息', default='')

    class Meta:
        unique_together = ('tool', 'project_id', 'cluster_id')

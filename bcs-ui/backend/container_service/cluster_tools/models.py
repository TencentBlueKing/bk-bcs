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
import re
from typing import Optional

from django.conf import settings
from django.db import models
from django.utils.translation import ugettext_lazy as _

from backend.helm.helm.providers.constants import PUBLIC_REPO_URL
from backend.utils.models import BaseModel

from .constants import ToolStatus


class Tool(models.Model):
    """组件库中的组件信息(通过 Helm Chart 管理)"""

    chart_name = models.CharField(max_length=128, unique=True)
    name = models.CharField(_('组件名'), max_length=64)
    default_version = models.CharField(_('单个组件的默认版本'), max_length=64, null=True, blank=True)
    default_values = models.TextField(null=True, blank=True, help_text=_('组件启用时需要额外设置的变量值，文本内容格式为 yaml'))
    # 记录一些额外的启动命令如 --disable-openapi-validation 等
    extra_options = models.TextField(default='')
    namespace = models.CharField(max_length=64, default='bcs-system')
    supported_actions = models.CharField(_('组件支持的操作'), max_length=128, default='install')
    description = models.TextField(help_text=_('组件功能介绍'), null=True, blank=True)
    help_link = models.CharField(max_length=255, null=True, blank=True)
    logo = models.TextField(_('图片 logo'), null=True, blank=True)
    version = models.CharField(_('组件库的版本'), max_length=64)

    @property
    def default_chart_url(self) -> str:
        if not self.default_version:
            return ''

        chart_tgz = f'{self.chart_name}-{self.default_version}.tgz'
        if settings.CLUSTER_TOOLS_REPO_PREFIX:
            return f'{settings.CLUSTER_TOOLS_REPO_PREFIX}/{chart_tgz}'
        else:
            return f'{PUBLIC_REPO_URL}/charts/{chart_tgz}'


class InstalledTool(BaseModel):
    """记录已安装到集群中的组件信息"""

    tool = models.ForeignKey(Tool, on_delete=models.CASCADE, db_constraint=False)
    # release names are limited to 53 characters.
    # https://helm.sh/docs/chart_template_guide/getting_started/#adding-a-simple-template-call
    release_name = models.CharField(max_length=53)
    project_id = models.CharField(max_length=32)
    cluster_id = models.CharField(max_length=32)
    chart_url = models.CharField(max_length=255)
    values = models.TextField(null=True, blank=True, help_text=_('组件启用或更新时设置的变量值，文本内容格式为 yaml'))
    extra_options = models.TextField(default='')
    namespace = models.CharField(max_length=64)
    status = models.CharField(choices=ToolStatus.get_choices(), default=ToolStatus.PENDING, max_length=32)
    message = models.TextField(_('记录错误信息'), default='')

    class Meta:
        unique_together = ('tool', 'project_id', 'cluster_id')

    @classmethod
    def create(
        cls, username: str, tool: Tool, project_id: str, cluster_id: str, values: Optional[str] = None
    ) -> 'InstalledTool':
        obj, _ = cls.objects.get_or_create(
            tool=tool,
            project_id=project_id,
            cluster_id=cluster_id,
            defaults={
                'release_name': tool.name.lower(),
                'chart_url': tool.default_chart_url,
                'values': values or tool.default_values,
                'extra_options': tool.extra_options,
                'namespace': tool.namespace,
                'creator': username,
                'updator': username,
            },
        )
        return obj

    @property
    def chart_version(self):
        """从 chart url 计算出 chart version"""
        chart_pkg = self.chart_url.rpartition('/')[-1]
        chart_name = self.tool.chart_name
        return re.sub(r'{}-(.*).tgz'.format(chart_name), r'\1', chart_pkg)

    def success(self):
        """安装或更新成功"""
        self._update_status(ToolStatus.DEPLOYED, 'success')

    def fail(self, err_msg: str):
        """变更失败. 变更包括安装, 更新和卸载"""
        self._update_status(ToolStatus.FAILED, err_msg)

    def on_upgrade(self, operator: str, chart_url: str, values: Optional[str] = None):
        """更新中的状态流转"""
        self.extra_options = self.tool.extra_options
        self.updator = operator
        self.chart_url = chart_url
        self.status = ToolStatus.PENDING
        self.message = 'start to upgrade'
        if values:
            self.values = values
        self.save()

    def on_delete(self, operator: str):
        """删除中的状态流转"""
        self.updator = operator
        self.status = ToolStatus.PENDING
        self.message = 'start to uninstall'
        self.save()

    def _update_status(self, status: str, message: str):
        self.status = status
        self.message = message
        self.save(update_fields=['status', 'message'])

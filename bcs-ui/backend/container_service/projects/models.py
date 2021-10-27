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
from django.utils import timezone
from django.utils.translation import ugettext_lazy as _

from backend.container_service.projects.constants import StaffInfoStatus
from backend.utils.models import BaseModel


class ProjectUser(models.Model):
    project_id = models.CharField(max_length=64)
    user_id = models.CharField(max_length=32)
    department = models.CharField(max_length=128, null=True, blank=True)

    joined_at = models.DateTimeField()
    leave_at = models.DateTimeField(null=True, blank=True)
    status = models.IntegerField(default=StaffInfoStatus.NORMAL.value, choices=StaffInfoStatus.get_choices())

    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)
    extra = models.TextField(null=True, blank=True)

    class Meta:
        unique_together = ('project_id', 'user_id')

    def __unicode__(self):
        return u'<{project_id},{user_id}({joined_at})>'.format(
            project_id=self.project_id, user_id=self.user_id, joined_at=self.joined_at
        )

    def to_dict(self):
        joined_at = timezone.localtime(self.joined_at).strftime('%Y-%m-%d')
        data = {
            'status': self.status,
            'status_desc': self.get_status_display(),
            'joined_at': joined_at,
            'department': self.department or '-',
        }
        return data


class DataID(BaseModel):
    project_id = models.CharField(_("项目ID"), max_length=32)
    data_id = models.IntegerField(_("数据平台DataID"))


class FunctionController(BaseModel):
    """
    功能开启控制器
    """

    func_code = models.CharField(_("功能code"), max_length=64, unique=True)
    func_name = models.CharField(_("功能名称"), max_length=64)
    enabled = models.BooleanField(_("是否开启该功能"), help_text=_("控制功能是否对外开放，若选择，则该功能将对外开放"), default=False)
    wlist = models.TextField(_("功能测试白名单"), blank=True, null=True, help_text=_("白名单，英文分号【;】隔开"))

    def __str__(self):
        return self.func_name

    class Meta:
        verbose_name = _('平台功能控制器')
        verbose_name_plural = _('平台功能控制器')


class Conf(BaseModel):
    """平台配置"""

    key = models.CharField(_("标识"), max_length=64, unique=True)
    name = models.CharField(_("名称"), max_length=128)
    value = models.TextField(_("值"))

    class Meta:
        verbose_name = _("平台配置")
        verbose_name_plural = _("平台配置")

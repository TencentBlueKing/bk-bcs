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


class BaseLoadBalance(BaseModel):
    project_id = models.CharField(_("项目ID"), max_length=32)
    cluster_id = models.CharField(_("集群ID"), max_length=32)
    namespace_id = models.IntegerField(default=0)
    name = models.CharField(_("名称"), max_length=32)

    class Meta:
        abstract = True


class K8SLoadBlance(BaseLoadBalance):
    protocol_type = models.CharField(_("协议类型"), max_length=32, default="https;http")
    ip_info = models.TextField()
    detail = models.TextField(help_text=_("可以用来存储整个配置信息"))
    namespace = models.CharField("命名空间名称", max_length=253, null=True, blank=True)

    class Meta:
        db_table = "k8s_load_blance"
        unique_together = ("cluster_id", "namespace", "name")


class MesosLoadBlance(BaseLoadBalance):
    linked_namespace_ids = models.TextField(null=True, blank=True)
    ip_list = models.TextField()
    data = models.TextField()
    data_dict = models.TextField(help_text=_("兼容历史数据,用以存储dict类型数据"))
    status = models.CharField(_("状态"), max_length=16, null=True, blank=True)
    namespace = models.CharField(_("命名空间名称"), max_length=32, null=True, blank=True)

    def update_status(self, status):
        self.status = status
        self.save(update_fields=["status"])

    class Meta:
        db_table = "mesos_load_blance"
        unique_together = ("cluster_id", "namespace_id", "name")

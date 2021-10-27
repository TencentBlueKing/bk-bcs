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

from .base import BaseModel
from .mixins import MConfigMapAndSecretMixin, PodMixin, ResourceMixin


class MesosResource(BaseModel):
    config = models.TextField(u"配置信息")
    name = models.CharField(u"名称", max_length=255, default='')

    class Meta:
        abstract = True
        ordering = ('created',)


class ConfigMap(MesosResource, MConfigMapAndSecretMixin):
    """
    Application ->挂在卷/环境变量 ->ConfigMap
    """


class Secret(MesosResource, MConfigMapAndSecretMixin):
    """
    Application ->挂在卷/环境变量 ->Secret
    """


class HPA(MesosResource, MConfigMapAndSecretMixin):
    """HPA数据表"""


class Application(MesosResource, PodMixin):
    """"""

    desc = models.TextField("描述", help_text="前台展示字段，bcs api 中无该信息")
    config = models.TextField("配置信息", help_text="包含：实例数量\ restart策略\kill策略\备注\调度约束\网络\容器信息")
    app_id = models.CharField("应用ID", max_length=32, help_text="每次保存时会生成新的应用记录，用app_id来记录与其他资源的关联关系")


class Deplpyment(MesosResource, ResourceMixin):
    """
    Deplpyment 是基于 Application 构建的
    本表只存储 Deplpyment 本身特性的相关策略

    {
        "app_id": "",
        "desc": "this is a desc",
        "config": {
            "strategy": {
                "type": "RollingUpdate",
                "rollingupdate": {
                    "maxUnavilable": 1,
                    "maxSurge": 1,
                    "upgradeDuration": 60,
                    "autoUpgrade": false,
                    "rollingOrder": "CreateFirst"
                }
            },
            "pause": false,
        }
    }
    """

    name = models.CharField("名称", max_length=255)
    app_id = models.CharField("关联的Application ID", max_length=32)
    desc = models.TextField("描述", help_text="前台展示字段，bcs api 中无该信息")
    config = models.TextField("配置信息", help_text="包含：升级策略")


class Service(MesosResource, ResourceMixin):
    """
    service 中的 ports 一定是在 application 中有的。application(ports) 大于等于 service(ports)
    """

    app_id = models.TextField("关联的Application ID", help_text="可以关联多个Application")


class Ingress(MesosResource, MConfigMapAndSecretMixin):
    """mesos ingress表"""

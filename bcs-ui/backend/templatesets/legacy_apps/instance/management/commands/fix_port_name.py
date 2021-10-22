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

Usage: python manage.py fix_port_name

将 Application中的 portname 字段修改为 portName
"""
from django.core.management.base import BaseCommand

from backend.templatesets.legacy_apps.configuration.models import Application


class Command(BaseCommand):
    help = u"Change portname to portName in Application healthChecks"

    def handle(self, *args, **options):
        old_name = '"portname":'
        new_name = '"portName":'

        apps = Application.objects.all()
        for _app in apps:
            _config = _app.config
            # 替换项目镜像的路径
            _config = _config.replace(old_name, new_name)
            _app.config = _config
            _app.updator = 'admin0123'
            _app.save()
        self.stdout.write(self.style.SUCCESS("Finish"))

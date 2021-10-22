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

Usage: python manage.py init_latest_ver
"""
import json

from django.core.management.base import BaseCommand

from backend.templatesets.legacy_apps.configuration.models import ShowVersion, Template, VersionedEntity


class Command(BaseCommand):
    help = u"将模板的最新版本初始化到可见版本中"

    def handle(self, *args, **options):
        all_temps = Template.objects.all()
        for tem in all_temps:
            # 将最新版本初始化到可见版本中
            last_version = VersionedEntity.objects.get_latest_by_template(tem.id)
            if last_version:
                is_show_ver = ShowVersion.objects.filter(template_id=tem.id, real_version_id=last_version.id).first()
                if is_show_ver:
                    self.stdout.write(
                        self.style.NOTICE(
                            'Already exist template_id[%s] show_version_id[%s] version_id[%s] name[%s]'
                            % (tem.id, is_show_ver.id, last_version.id, is_show_ver.name)
                        )
                    )
                else:
                    show_ver = ShowVersion.objects.create(
                        template_id=tem.id,
                        real_version_id=last_version.id,
                        name=last_version.version,
                        history=json.dumps([last_version.id]),
                    )
                    self.stdout.write(
                        self.style.SUCCESS(
                            'Successfully initialize template_id[%s] show_version_id[%s] version_id[%s] name[%s]'
                            % (tem.id, show_ver.id, last_version.id, show_ver.name)
                        )
                    )

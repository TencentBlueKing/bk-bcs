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

Usage: python manage.py init_version
"""
import json

from django.core.management.base import BaseCommand

from backend.templatesets.legacy_apps.configuration.models import ShowVersion, VersionedEntity
from backend.templatesets.legacy_apps.instance.models import VersionInstance


class Command(BaseCommand):
    help = u"Initialize the instantiated version into the show version table"

    def handle(self, *args, **options):
        init_ins_sets = VersionInstance.objects.filter(show_version_id=0)
        # init_ins_sets = VersionInstance.objects.all()
        for ins in init_ins_sets:
            version_id = ins.version_id
            template_id = ins.template_id
            show_ver = ShowVersion.default_objects.filter(real_version_id=version_id, template_id=template_id).first()
            if not show_ver:
                version_entity = VersionedEntity.objects.get(id=version_id)
                show_ver = ShowVersion.objects.create(
                    template_id=template_id,
                    real_version_id=version_id,
                    name=version_entity.version,
                    history=json.dumps([version_id]),
                )
            ins.show_version_id = show_ver.id
            ins.show_version_name = show_ver.name
            ins.save()
            self.stdout.write(
                self.style.SUCCESS(
                    'Successfully initialize template_id[%s] version_id[%s] name[%s]'
                    % (template_id, version_id, show_ver.name)
                )
            )

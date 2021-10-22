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

给前端同学伪造模板集中镜像不存在的情况，修复相关体验的bug
Usage: python manage.py fake_image_path
"""

from django.core.management.base import BaseCommand

from backend.templatesets.legacy_apps.configuration.models import Application, Template, VersionedEntity


class Command(BaseCommand):
    help = u"Initialize the image path in the template"

    def handle(self, *args, **options):
        pro_dic = {'2859855c38ce49e3a8755b499ef0b433': 'a88'}

        for project_id in pro_dic:
            old_public_path = 'paas_test/public/'
            new_public_path = 'paas/%s/' % project_id

            temps = Template.objects.filter(project_id=project_id)
            app_id_list = []
            for tem in temps:
                vers = VersionedEntity.objects.filter(template_id=tem.id)
                for _ver in vers:
                    _entity = _ver.get_entity()
                    app_ids = _entity.get('application')
                    id_list = app_ids.split(',') if app_ids else []
                    app_id_list.extend(id_list)

            applications = Application.objects.filter(id__in=app_id_list)
            for _app in applications:
                _config = _app.config
                # 替换项目镜像的路径
                _config = _config.replace(old_public_path, new_public_path)
                _app.config = _config
                _app.updator = 'admin'
                _app.save()

            if len(app_id_list) > 0:
                self.stdout.write(
                    self.style.SUCCESS(
                        '%s[%s]: [%s] -> [%s]' % (project_id, len(app_id_list), old_public_path, new_public_path)
                    )
                )

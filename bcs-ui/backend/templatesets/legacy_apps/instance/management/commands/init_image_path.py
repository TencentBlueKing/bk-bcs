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

Usage: python manage.py init_image_path -n 'IAnlbGg31xLXHdfEYOjpYiT4toebWH'

paas/01b6ad17aafc49dcb5cb1aa3d6ee6e01/ -> paas_test/job/,paas/job/

paas/public/ -> paas_test/public/   : 测试环境
"""
import requests
from django.conf import settings
from django.core.management.base import BaseCommand

from backend.templatesets.legacy_apps.configuration.models import Application, Template, VersionedEntity


class Command(BaseCommand):
    help = u"Initialize the image path in the template"

    def add_arguments(self, parser):

        parser.add_argument(
            '-n',
            '--access_token',
            action='store',
            dest='access_token',
            default='',
            help='access_token',
        )

    def handle(self, *args, **options):
        print(options)
        if options.get('access_token'):
            access_token = options['access_token']
        else:
            return False
        # 获取所有项目的id 和 英文名
        cc_host = '%s/api/paas-cc/%s' % (settings.APIGW_HOST, settings.APIGW_ENV)
        url = '{host}/project/get_project_list/'.format(**{'host': cc_host})
        params = {'access_token': access_token}
        project_res = requests.request('get', url, params=params).json()

        pro_data = project_res.get('data') or []
        pro_dic = {}
        for _d in pro_data:
            pro_dic[_d.get('project_id')] = _d.get('english_name')

        for project_id in pro_dic:
            english_name = pro_dic[project_id]
            old_jfrog_path = 'paas/%s/' % project_id
            new_jfrog_path = '%s/%s/' % (settings.DEPOT_PREFIX, english_name)

            old_public_path = 'paas/public/'
            new_public_path = '%s/public/' % settings.DEPOT_PREFIX

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
                _config = _config.replace(old_jfrog_path, new_jfrog_path)
                _config = _config.replace(old_public_path, new_public_path)
                _app.config = _config
                _app.updator = 'admin0115'
                _app.save()

            if len(app_id_list) > 0:
                self.stdout.write(
                    self.style.SUCCESS(
                        '%s[%s]: [%s] -> [%s]' % (project_id, len(app_id_list), old_jfrog_path, new_jfrog_path)
                    )
                )
                self.stdout.write(
                    self.style.SUCCESS('%s: [%s] -> [%s]' % (project_id, old_public_path, new_public_path))
                )

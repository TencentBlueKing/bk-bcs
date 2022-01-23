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

针对k8s项目检查是否注册公共仓库，如果没有注册，注册相应项目的公共仓库信息
"""

from django.core.management.base import BaseCommand
from django.utils.translation import ugettext_lazy as _

from backend.accounts import bcs_perm
from backend.components import paas_cc
from backend.helm.helm.providers import repo_provider


class Command(BaseCommand):
    def get_client_access_token(self):
        access_token = bcs_perm.get_access_token().get('access_token')
        if not access_token:
            self.stdout.write(_("获取access_token失败"))
            return
        return access_token

    def get_all_projects(self, access_token):
        """获取所有项目信息"""
        projects = paas_cc.get_projects(access_token, query_params={'desire_all_data': 1})
        if projects.get('code') != 0:
            self.stdout.write(projects.get('message'))
            return
        return projects.get('data') or []

    def handle(self, *args, **options):
        access_token = self.get_client_access_token()
        if not access_token:
            return

        projects = self.get_all_projects(access_token)
        if not projects:
            return
        project_id_list = [info['project_id'] for info in projects if info['kind'] == 1 and info['is_offlined'] == 0]
        # 注册db记录
        try:
            for project_id in project_id_list:
                repo_provider.add_platform_public_repos(project_id)
        except Exception as err:
            self.stdout.write(f"{_('创建项目公共chart失败，详细信息')}: {err}")

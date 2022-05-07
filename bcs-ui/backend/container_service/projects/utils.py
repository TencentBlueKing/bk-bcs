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
import logging

from celery import shared_task
from django.utils.translation import ugettext_lazy as _

from backend.components import bk_repo, cc
from backend.container_service.projects.drivers import K8SDriver
from backend.templatesets.legacy_apps.configuration.init_data import init_template
from backend.utils import FancyDict
from backend.utils.notify import notify_manager

logger = logging.getLogger(__name__)


def fetch_has_maintain_perm_apps(request):
    return cc.fetch_has_maintain_perm_apps(request.user.username)


def create_bkrepo_project_and_depot(project: FancyDict, username: str):
    """创建制品库项目及镜像仓库"""
    client = bk_repo.BkRepoClient()
    # 创建项目
    try:
        client.create_project(project.project_code, project.project_name, project.description)
    except bk_repo.BkRepoCreateProjectError as e:
        logger.error("创建制品库的项目失败: %s", e)
    # 创建镜像仓库
    try:
        client.create_repo(f"{project.project_code}-docker", repo_type="DOCKER")
    except bk_repo.BkRepoCreateRepoError as e:
        logger.error("创建制品库的镜像仓库失败: %s", e)


def update_bcs_service_for_project(request, project_id, data):
    username = request.user.username
    create_bkrepo_project_and_depot(request.project, username)
    logger.info(f'init_template [update] project_id: {project_id}')
    init_template.delay(
        project_id, request.project.english_name, request.user.token.access_token, request.user.username
    )
    # helm handler
    K8SDriver.backend_create_helm_info(project_id)
    notify_manager.delay(_("用户[{}]在项目{}下启用了容器服务，请关注").format(username, request.project.project_name))


try:
    from .utils_ext import *  # noqa
except ImportError as e:
    logger.debug("Load extension failed: %s", e)

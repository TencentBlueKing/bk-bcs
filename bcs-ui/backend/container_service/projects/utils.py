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

from django.utils.translation import ugettext_lazy as _

from backend.components import cc
from backend.container_service.misc.depot.api import create_project_path_by_api
from backend.container_service.projects.drivers import K8SDriver
from backend.templatesets.legacy_apps.configuration.init_data import init_template
from backend.utils.notify import notify_manager

logger = logging.getLogger(__name__)


def backend_create_depot_path(request, project_id, pre_cc_app_id):
    try:
        create_project_path_by_api(request.user.token.access_token, project_id, request.project.english_name)
    except Exception as err:
        logger.error("创建项目仓库路径失败，详细信息: %s" % err)


def fetch_has_maintain_perm_apps(request):
    return cc.fetch_has_maintain_perm_apps(request.user.username)


def update_bcs_service_for_project(request, project_id, data):
    backend_create_depot_path(request, project_id, request.project.cc_app_id)
    logger.info(f'init_template [update] project_id: {project_id}')
    init_template.delay(
        project_id, request.project.english_name, request.user.token.access_token, request.user.username
    )
    # helm handler
    K8SDriver.backend_create_helm_info(project_id)
    notify_manager.delay(
        '{prefix_msg}[{username}]{project}{project_name}{suffix_msg}'.format(
            prefix_msg=_("用户"),
            username=request.user.username,
            project=_("在项目"),
            project_name=request.project.project_name,
            suffix_msg=_("下启用了容器服务，请关注"),
        )
    )


try:
    from .utils_ext import *  # noqa
except ImportError as e:
    logger.debug("Load extension failed: %s", e)

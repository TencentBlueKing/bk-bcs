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
from typing import Optional

from rest_framework.permissions import BasePermission

from backend.bcs_web.audit_log.audit.context import AuditContext
from backend.components.base import ComponentAuth
from backend.components.paas_cc import PaaSCCClient
from backend.container_service.clusters.base.models import CtxCluster
from backend.container_service.projects.base.models import CtxProject
from backend.iam.permissions.resources.project import ProjectPermCtx, ProjectPermission
from backend.utils import FancyDict
from backend.utils.cache import region

from .constants import bcs_project_cache_key

EXPIRATION_TIME = 3600 * 24 * 30


class AccessProjectPermission(BasePermission):
    """仅支持处理 url 路径参数中包含 project_id 或 project_id_or_code 的接口"""

    message = "no project permissions"

    def has_permission(self, request, view):
        if request.user.is_superuser:
            return True

        access_token = request.user.token.access_token

        project_id_or_code = view.kwargs.get('project_id') or view.kwargs.get('project_id_or_code')
        project_id = self._get_project_id(access_token, project_id_or_code)
        if not project_id:
            return False

        perm_ctx = ProjectPermCtx(username=request.user.username, project_id=project_id)
        return ProjectPermission().can_view(perm_ctx, raise_exception=False)

    def _get_project_id(self, access_token, project_id_or_code: str) -> str:
        cache_key = f'BK_DEVOPS_BCS:PROJECT_ID:{project_id_or_code}'
        project_id = region.get(cache_key, expiration_time=EXPIRATION_TIME)

        if not project_id:
            paas_cc = PaaSCCClient(auth=ComponentAuth(access_token))
            project_data = paas_cc.get_project(project_id_or_code)
            project_id = project_data['project_id']
            region.set(cache_key, project_id)

        return project_id


class ProjectEnableBCS(BasePermission):
    """
    仅支持处理 url 路径参数中包含 project_id 或 project_id_or_code 的接口
    主要功能:
    - 校验项目是否已经开启容器服务
    - 设置 request.project、request.ctx_project 、request.ctx_cluster，在 view 中使用
    - 设置 request.audit_ctx，可配合 log_audit_on_view 和 log_audit 装饰器使用
    """

    message = "project does not enable bcs"

    def has_permission(self, request, view):
        project_id_or_code = view.kwargs.get('project_id') or view.kwargs.get('project_id_or_code')
        project = self._get_enabled_project(request.user.token.access_token, project_id_or_code)
        if project:
            request.project = project
            self._set_ctx_project_cluster(request, project.project_id, view.kwargs.get('cluster_id', ''))
            # 设置操作审计 context
            request.audit_ctx = AuditContext(user=request.user.username, project_id=project.project_id)
            return True

        return False

    def _get_enabled_project(self, access_token, project_id_or_code: str) -> Optional[FancyDict]:
        cache_key = bcs_project_cache_key.format(project_id_or_code=project_id_or_code)
        project = region.get(cache_key, expiration_time=EXPIRATION_TIME)
        if project and isinstance(project, FancyDict):
            return project

        paas_cc = PaaSCCClient(auth=ComponentAuth(access_token))
        project_data = paas_cc.get_project(project_id_or_code)
        project = FancyDict(**project_data)

        self._refine_project(project)

        # 项目绑定了业务，即开启容器服务
        if project.cc_app_id != 0:
            region.set(cache_key, project)
            return project

        return None

    def _refine_project(self, project: FancyDict):
        project.coes = project.kind
        project.project_code = project.english_name

        try:
            from backend.container_service.projects.utils import get_project_kind

            # k8s类型包含kind为1(bcs k8s)或其它属于k8s的编排引擎
            project.kind = get_project_kind(project.kind)
        except ImportError:
            pass

    def _set_ctx_project_cluster(self, request, project_id: str, cluster_id: str):
        access_token = request.user.token.access_token
        request.ctx_project = CtxProject.create(token=access_token, id=project_id)
        if cluster_id:
            request.ctx_cluster = CtxCluster.create(token=access_token, id=cluster_id, project_id=project_id)
        else:
            request.ctx_cluster = None

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
import operator

from django.conf import settings
from django.utils.translation import ugettext_lazy as _
from rest_framework import permissions, viewsets
from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.response import Response
from rest_framework.views import APIView

from backend.bcs_web.audit_log import client
from backend.bcs_web.constants import bcs_project_cache_key
from backend.bcs_web.viewsets import SystemViewSet
from backend.components import cc, paas_cc
from backend.container_service.projects import base as Project
from backend.container_service.projects.utils import fetch_has_maintain_perm_apps, update_bcs_service_for_project
from backend.iam.permissions.decorators import response_perms
from backend.iam.permissions.resources.project import (
    ProjectAction,
    ProjectCreatorAction,
    ProjectPermCtx,
    ProjectPermission,
    ProjectRequest,
)
from backend.utils.cache import region
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes
from backend.utils.renderers import BKAPIRenderer
from backend.utils.response import PermsResponse

from . import serializers
from .authorized import list_auth_projects
from .cmdb import list_biz_maintainers

logger = logging.getLogger(__name__)


class Projects(viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_project(self, request, project_id):
        """单个项目信息"""
        data = request.project
        # 添加业务名称
        data["cc_app_name"] = cc.get_application_name(request.project.cc_app_id)
        return Response(data)

    def update_bound_biz(self, request, project_id):
        """更新项目信息"""
        if not self._can_update_bound_biz(request, project_id):
            raise error_codes.CheckFailed(_("请确认有项目管理员权限，并且项目下无集群"))

        data = self._validate_update_project_data(request)
        access_token = request.user.token.access_token
        data["updator"] = request.user.username

        # 添加操作日志
        ual_client = client.UserActivityLogClient(
            project_id=project_id,
            user=request.user.username,
            resource_type="project",
            resource=request.project.project_name,
            resource_id=project_id,
            description="{}: {}".format(_("更新项目"), request.project.project_name),
        )
        resp = paas_cc.update_project_new(access_token, project_id, data)
        if resp.get("code") != ErrorCode.NoError:
            ual_client.log_modify(activity_status="failed")
            raise error_codes.APIError(_("更新项目信息失败，错误详情: {}").format(resp.get("message")))
        ual_client.log_modify(activity_status="succeed")

        project_data = resp.get("data")
        # 主动令缓存失效
        self._invalid_project_cache(project_id)
        # 创建或更新依赖服务，包含data、template、helm
        update_bcs_service_for_project(request, project_id, data)

        return Response(project_data)

    def _can_update_bound_biz(self, request, project_id):
        """判断是否允许修改项目
        - 项目下有集群，不允许更改项目的绑定业务
        - 非管理员权限，不允许修改项目
        """
        if self._has_cluster(request.user.token.access_token, project_id):
            return False

        perm_ctx = ProjectPermCtx(username=request.user.username, project_id=project_id)
        if not ProjectPermission().can_edit(perm_ctx, raise_exception=False):
            return False

        return True

    def _has_cluster(self, access_token: str, project_id: str):
        """判断项目下是否有集群"""
        resp = paas_cc.get_all_clusters(access_token, project_id)
        if resp.get("code") != ErrorCode.NoError:
            raise error_codes.APIError(resp.get("message"))
        # 存在集群时，不允许修改
        if resp.get("data", {}).get("count") > 0:
            return True
        return False

    def _validate_update_project_data(self, request):
        serializer = serializers.UpdateProjectNewSLZ(data=request.data)
        serializer.is_valid(raise_exception=True)
        return serializer.data

    def _invalid_project_cache(self, project_id):
        """当变更项目信息时，详细缓存信息失效"""
        region.delete(bcs_project_cache_key.format(project_id_or_code=project_id))
        # NOTE: 后续permission统一后，可以删除下面的缓存标识
        region.delete(f"BK_DEVOPS_BCS:HAS_BCS_SERVICE:{project_id}")


class CC(viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def list(self, request):
        """获取当前用户CC列表"""
        data = fetch_has_maintain_perm_apps(request)
        return Response(data)


class UserAPIView(APIView):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)
    permission_classes = (permissions.IsAuthenticated,)

    def get(self, request):
        data = {
            "avatar_url": f"{settings.BK_PAAS_HOST}/static/img/getheadimg.jpg",
            "username": request.user.username,
            "chinese_name": "",
            "permissions": [],
        }
        return Response(data)


class AuthorizedProjectsViewSet(viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)
    permission_classes = (permissions.IsAuthenticated,)

    def list(self, request):
        """查询用户有权限的项目列表"""
        resp = list_auth_projects(request.user.token.access_token, request.user.username)
        if resp.get('code') != ErrorCode.NoError:
            logger.error('list_auth_projects error: %s', resp.get('message'))
            raise error_codes.ComponentError('list auth projects error')

        projects = resp['data']
        if not projects:
            return Response([])

        projects.sort(key=operator.itemgetter('created_at'), reverse=True)
        return Response(projects)


class NavProjectsViewSet(viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)
    permission_classes = (permissions.IsAuthenticated,)
    iam_perm = ProjectPermission()

    def create_project(self, request):
        username = request.user.username

        perm_ctx = ProjectPermCtx(username=username)
        self.iam_perm.can_create(perm_ctx)

        req_data = request.data.copy()
        req_data["creator"] = username
        serializer = serializers.CreateNavProjectSLZ(data=req_data)
        serializer.is_valid(raise_exception=True)

        project = Project.create_project(request.user.token.access_token, serializer.validated_data)
        self.iam_perm.grant_resource_creator_actions(
            ProjectCreatorAction(name=project["project_name"], project_id=project["project_id"], creator=username),
        )

        return Response(project)

    def update_project(self, request, project_id):
        perm_ctx = ProjectPermCtx(username=request.user.username, project_id=project_id)
        self.iam_perm.can_edit(perm_ctx)

        req_data = request.data.copy()
        req_data["updator"] = request.user.username
        serializer = serializers.UpdateNavProjectSLZ(data=req_data)
        serializer.is_valid(raise_exception=True)

        project = Project.update_project(request.user.token.access_token, project_id, serializer.validated_data)
        return Response(project)

    @response_perms(
        action_ids=[ProjectAction.VIEW, ProjectAction.EDIT],
        permission_cls=ProjectPermission,
        resource_id_key='project_id',
    )
    def list_projects(self, request):
        """
        提供查询项目列表的功能, 同时能够根据 project_code 或 project_name 过滤项目
        """
        project_code = request.query_params.get("project_code")
        project_name = request.query_params.get("project_name")
        access_token = request.user.token.access_token

        if project_code:
            projects = Project.list_projects(access_token, {"english_names": project_code})
        elif project_name:
            projects = Project.list_projects(access_token, {"project_names": project_name})
        else:
            projects = Project.list_projects(access_token)

        if not projects:
            return Response(projects)

        projects.sort(key=operator.itemgetter('created_at'), reverse=True)
        return PermsResponse(projects, ProjectRequest())

    @response_perms(
        action_ids=[ProjectAction.VIEW, ProjectAction.EDIT],
        permission_cls=ProjectPermission,
        resource_id_key='project_id',
    )
    def get_project(self, request, project_id):
        project = Project.get_project(request.user.token.access_token, project_id)
        return PermsResponse(project, ProjectRequest())


class ProjectBizInfoViewSet(SystemViewSet):
    def list_biz_maintainers(self, request, project_id):
        """查询业务下的运维人员"""
        # 以admin身份查询业务下的运维
        maintainers = list_biz_maintainers(int(request.project.cc_app_id))
        return Response({"maintainers": maintainers})

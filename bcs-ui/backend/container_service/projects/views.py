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
import json
import logging

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
from backend.iam.legacy_perms import ProjectPermission
from backend.utils.basic import normalize_datetime
from backend.utils.cache import region
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes
from backend.utils.renderers import BKAPIRenderer

from . import serializers
from .cmdb import list_biz_maintainers

logger = logging.getLogger(__name__)


class Projects(viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def normalize_create_update_time(self, created_at, updated_at):
        return normalize_datetime(created_at), normalize_datetime(updated_at)

    def deploy_type_list(self, deploy_type):
        """转换deploy_type为list类型"""
        if not deploy_type:
            return []
        if str.isdigit(str(deploy_type)):
            deploy_type_list = [int(deploy_type)]
        else:
            try:
                deploy_type_list = json.loads(deploy_type)
            except Exception as err:
                logger.error("解析部署类型失败，详情: %s", err)
                return []
        return deploy_type_list

    def list(self, request):
        """获取项目列表"""
        # 获取已经授权的项目
        access_token = request.user.token.access_token
        # 直接调用配置中心接口去获取信息
        projects = paas_cc.get_auth_project(access_token)
        if projects.get("code") != ErrorCode.NoError:
            raise error_codes.APIError(projects.get("message"))
        data = projects.get("data")
        # 兼容先前，返回array/list
        if not data:
            return Response([])
        # 按数据倒序排序
        data.sort(key=lambda x: x["created_at"], reverse=True)
        # 数据处理
        for info in data:
            info["created_at"], info["updated_at"] = self.normalize_create_update_time(
                info["created_at"], info["updated_at"]
            )
            info["project_code"] = info["english_name"]
            info["deploy_type"] = self.deploy_type_list(info.get("deploy_type"))

        return Response(data)

    def has_cluster(self, request, project_id):
        """判断项目下是否有集群"""
        resp = paas_cc.get_all_clusters(request.user.token.access_token, project_id)
        if resp.get("code") != ErrorCode.NoError:
            raise error_codes.APIError(resp.get("message"))
        # 存在集群时，不允许修改
        if resp.get("data", {}).get("count") > 0:
            return True
        return False

    def can_edit(self, request, project_id):
        """判断是否允许修改项目
        - 项目下有集群，不允许更改项目的调度类型和绑定业务
        - 非管理员权限，不允许修改项目
        """
        perm = ProjectPermission()
        if (self.has_cluster(request, project_id)) or (not perm.can_edit(request.user.username, project_id)):
            return False
        return True

    def info(self, request, project_id):
        """单个项目信息"""
        data = request.project
        data["created_at"], data["updated_at"] = self.normalize_create_update_time(
            data["created_at"], data["updated_at"]
        )
        # 添加业务名称
        data["cc_app_name"] = cc.get_application_name(request.project.cc_app_id)
        data["can_edit"] = self.can_edit(request, project_id)
        return Response(data)

    def validate_update_project_data(self, request):
        serializer = serializers.UpdateProjectNewSLZ(data=request.data)
        serializer.is_valid(raise_exception=True)
        return serializer.data

    def invalid_project_cache(self, project_id):
        """当变更项目信息时，详细缓存信息失效"""
        region.delete(bcs_project_cache_key.format(project_id_or_code=project_id))
        # NOTE: 后续permission统一后，可以删除下面的缓存标识
        region.delete(f"BK_DEVOPS_BCS:HAS_BCS_SERVICE:{project_id}")

    def update(self, request, project_id):
        """更新项目信息"""
        if not self.can_edit(request, project_id):
            raise error_codes.CheckFailed(_("请确认有项目管理员权限，并且项目下无集群"))
        data = self.validate_update_project_data(request)
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
        if project_data:
            project_data["created_at"], project_data["updated_at"] = self.normalize_create_update_time(
                project_data["created_at"], project_data["updated_at"]
            )

        # 主动令缓存失效
        self.invalid_project_cache(project_id)
        # 创建或更新依赖服务，包含data、template、helm
        update_bcs_service_for_project(request, project_id, data)

        return Response(project_data)


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


class NavProjectsViewSet(viewsets.ViewSet, ProjectPermission):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)
    permission_classes = (permissions.IsAuthenticated,)

    def create_project(self, request):
        username = request.user.username
        self.can_create(username, raise_exception=True)

        req_data = request.data.copy()
        req_data["creator"] = username
        serializer = serializers.CreateNavProjectSLZ(data=req_data)
        serializer.is_valid(raise_exception=True)

        project = Project.create_project(request.user.token.access_token, serializer.validated_data)
        self.grant_related_action_perms(username, project["project_id"], project["project_name"])
        return Response(project)

    def update_project(self, request, project_id):
        self.can_edit(request.user.username, project_id, raise_exception=True)

        req_data = request.data.copy()
        req_data["updator"] = request.user.username
        serializer = serializers.UpdateNavProjectSLZ(data=req_data)
        serializer.is_valid(raise_exception=True)

        project = Project.update_project(request.user.token.access_token, project_id, serializer.validated_data)
        return Response(project)

    def _add_permissions_field(self, projects, username):
        project_ids = [p["project_id"] for p in projects]
        resource_perm_allowed = self.batch_resource_multi_actions_allowed(
            username, [self.actions.VIEW.value, self.actions.EDIT.value], project_ids
        )
        for p in projects:
            p["permissions"] = resource_perm_allowed[p["project_id"]]

    def filter_projects(self, request):
        project_code = request.query_params.get("project_code")
        project_name = request.query_params.get("project_name")
        with_permissions_field = request.query_params.get("with_permissions_field")
        access_token = request.user.token.access_token

        if project_code:
            projects = Project.list_projects(access_token, {"english_names": project_code})
        elif project_name:
            projects = Project.list_projects(access_token, {"project_names": project_name})
        else:
            projects = Project.list_projects(access_token)

        if not projects:
            return Response(projects)

        if with_permissions_field != "false":  # 需要权限字段
            self._add_permissions_field(projects, request.user.username)

        projects.sort(key=lambda p: p.get("created_at", ""), reverse=True)
        return Response(projects)

    def get_project(self, request, project_id):
        project = Project.get_project(request.user.token.access_token, project_id)
        project["permissions"] = self.resource_inst_multi_actions_allowed(
            request.user.username, [self.actions.VIEW.value, self.actions.EDIT.value], project_id
        )
        return Response(project)


class NavProjectPermissionViewSet(viewsets.ViewSet, ProjectPermission):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)
    permission_classes = (permissions.IsAuthenticated,)

    def get_user_perms(self, request):
        serializer = serializers.ProjectPermsSLZ(data=request.data)
        serializer.is_valid(raise_exception=True)
        perms = self.query_user_perms(request.user.username, **serializer.validated_data)
        return Response(perms)

    def query_user_perms_by_project(self, request, project_id):
        req_data = request.data.copy()
        req_data["project_id"] = project_id

        serializer = serializers.ProjectInstPermsSLZ(data=req_data)
        serializer.is_valid(raise_exception=True)

        perms = self.query_user_perms(request.user.username, **serializer.validated_data)
        return Response(perms)

    def list_authorized_users(self, request, project_id):
        serializer = serializers.QueryAuthorizedUsersSLZ(data=request.query_params)
        serializer.is_valid(raise_exception=True)

        users = self.query_authorized_users(project_id, serializer.validated_data["action_id"])
        return Response(users)


class ProjectBizInfoViewSet(SystemViewSet):
    def list_biz_maintainers(self, request, project_id):
        """查询业务下的运维人员"""
        # 以admin身份查询业务下的运维
        maintainers = list_biz_maintainers(int(request.project.cc_app_id))
        return Response({"maintainers": maintainers})


# TODO: 是否有其它方式处理
try:
    from .views_ext import patch_project_client

    Projects = patch_project_client(Projects)
except ImportError as e:
    logger.debug("Load extension failed: %s", e)

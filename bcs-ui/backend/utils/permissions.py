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
from rest_framework.permissions import BasePermission

from backend.components import paas_cc
from backend.iam.permissions.resources import ProjectPermCtx, ProjectPermission
from backend.utils import FancyDict
from backend.utils.cache import region
from backend.utils.error_codes import error_codes

# 跳过路径为projects、projects_pub的namespace的项目校验
SKIP_REQUEST_NAMESPACE = ["projects", "projects_pub"]


class HasProject(BasePermission):
    def has_permission(self, request, view):
        project_id = view.kwargs.get("project_id")
        if not project_id:
            return True

        if request.user.is_superuser:
            return True

        perm_ctx = ProjectPermCtx(username=request.user.username, project_id=project_id)
        return ProjectPermission().can_view(perm_ctx, raise_exception=False)


class HasIAMProject(BasePermission):
    def has_permission(self, request, view):
        project_id = view.kwargs.get("project_id")
        if not project_id:
            return True

        if request.user.is_superuser:
            return True

        access_token = request.user.token.access_token

        project_code = self.get_project_code(access_token, project_id)
        if not project_code:
            return False

        perm_ctx = ProjectPermCtx(
            username=request.user.username, project_id=self.get_project_id(access_token, project_id)
        )
        return ProjectPermission().can_view(perm_ctx, raise_exception=False)

    def get_project_code(self, access_token, project_id):
        """获取project_code
        缓存较长时间
        """
        cache_key = f"BK_DEVOPS_BCS:PROJECT_CODE:{project_id}"
        project_code = region.get(cache_key, expiration_time=3600 * 24 * 30)
        if not project_code:
            # 这里的project_id对应实际的project_id或project_code, paas_cc接口兼容了两种类型的查询
            result = paas_cc.get_project(access_token, project_id)
            if result.get("code") != 0:
                return None
            project_code = result["data"]["english_name"]
            region.set(cache_key, project_code)
        return project_code

    def get_project_id(self, access_token, project_id):
        """获取project_id
        缓存较长时间
        # TODO 临时使用
        """
        cache_key = f"BK_DEVOPS_BCS:REAL_PROJECT_ID:{project_id}"
        real_project_id = region.get(cache_key, expiration_time=3600 * 24 * 30)
        if not real_project_id:
            # 这里的project_id对应实际的project_id或project_code, paas_cc接口兼容了两种类型的查询
            result = paas_cc.get_project(access_token, project_id)
            if result.get("code") != 0:
                return None
            real_project_id = result["data"]["project_id"]
            region.set(cache_key, real_project_id)
        return real_project_id


class ProjectHasBCS(BasePermission):
    def has_permission(self, request, view):
        project_id = view.kwargs.get("project_id")
        if not project_id:
            return True

        access_token = request.user.token.access_token

        request_namespace = request.resolver_match.namespace
        project = self.has_bcs_service(access_token, project_id, request_namespace)

        # 赋值给request.project
        project["project_code"] = project["english_name"]
        request.project = project

        return True

    def has_bcs_service(self, access_token, project_id, request_namespace):
        """判断是否开启容器服务
        开启后就不能关闭，所以缓存很久，默认30天
        """
        cache_key = f"BK_DEVOPS_BCS:HAS_BCS_SERVICE:{project_id}"
        project = region.get(cache_key, expiration_time=3600 * 24 * 30)

        if not project or not isinstance(project, FancyDict):
            result = paas_cc.get_project(access_token, project_id)
            project = result.get("data") or {}

            # coes: container orchestration engines
            project['coes'] = project['kind']
            try:
                from backend.container_service.projects.utils import get_project_kind

                # k8s类型包含kind为1(bcs k8s)或其它属于k8s的编排引擎
                project['kind'] = get_project_kind(project['kind'])
            except ImportError:
                pass

            project = FancyDict(project)

            if request_namespace in SKIP_REQUEST_NAMESPACE:
                # 如果是SKIP_REQUEST_NAMESPACE，有更新接口，不判断kind
                if project.get("cc_app_id") != 0:
                    region.set(cache_key, project)

            elif project.get("cc_app_id") != 0:
                region.set(cache_key, project)
            else:
                # 其他抛出没有开启容器服务
                raise error_codes.NoBCSService()

        return project

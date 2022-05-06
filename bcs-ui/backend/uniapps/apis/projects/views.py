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
from django.http import JsonResponse

from backend.container_service.projects.authorized import list_auth_projects
from backend.uniapps.apis.base_views import BaseAPIViews
from backend.uniapps.apis.projects import serializers
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes


class ProjectList(BaseAPIViews):
    def get(self, request, cc_app_id):
        """获取项目列表
        关联CC业务ID的，并且用户有权限的项目
        """
        params = request.query_params
        params_slz = serializers.ProjectListParamsSLZ(data=params)
        params_slz.is_valid(raise_exception=True)
        params_slz = params_slz.data
        project_info = list_auth_projects(params_slz["access_token"])
        # 通过cc app id过滤
        if project_info.get("code") != ErrorCode.NoError:
            raise error_codes.APIError.f(project_info.get("message"))
        ret_data = [info for info in project_info.get("data") or [] if str(info.get("cc_app_id")) == str(cc_app_id)]
        return JsonResponse({"code": 0, "data": ret_data})

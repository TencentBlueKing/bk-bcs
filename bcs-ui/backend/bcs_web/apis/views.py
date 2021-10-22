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
from django.utils.translation import ugettext_lazy as _
from rest_framework import viewsets

from backend.components import paas_cc

try:
    from backend.components.paas_auth_ext import get_access_token as get_client_access_token
except ImportError:
    from backend.components.ssm import get_client_access_token

from backend.utils import FancyDict
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes
from backend.utils.permissions import HasIAMProject, ProjectHasBCS
from backend.utils.renderers import BKAPIRenderer

from .authentication import JWTAuthentication
from .permissions import AccessTokenPermission, RemoteAccessPermission


class BaseAPIViewSet(viewsets.ViewSet):
    authentication_classes = (JWTAuthentication,)
    permission_classes = (AccessTokenPermission, HasIAMProject, ProjectHasBCS)
    renderer_classes = (BKAPIRenderer,)

    def initial(self, request, *args, **kwargs):
        if "project_id" in kwargs:
            return super().initial(request, *args, **kwargs)

        self.kwargs["project_id"] = kwargs.get("project_id_or_code") or kwargs.get("project_code")
        super().initial(request, *args, **kwargs)
        del self.kwargs["project_id"]


class NoAccessTokenBaseAPIViewSet(BaseAPIViewSet):
    permission_classes = (RemoteAccessPermission, HasIAMProject, ProjectHasBCS)

    def initial(self, request, *args, **kwargs):
        request.user.token = FancyDict(access_token=get_client_access_token().get("access_token"))
        super().initial(request, *args, **kwargs)


class ProjectBaseAPIViewSet(viewsets.ViewSet):
    """对流水线等外部调用API URL不定参数, 转换为内部的project_id, project_code等"""

    authentication_classes = (JWTAuthentication,)
    permission_classes = (RemoteAccessPermission, HasIAMProject, ProjectHasBCS)
    renderer_classes = (BKAPIRenderer,)

    # 具体view函数需要的字段名称
    project_field_name = "project_id"
    available_project_field_names = ["project_id", "project_code", "project_id_or_code"]

    def initial(self, request, *args, **kwargs):
        request.user.token = FancyDict(access_token=get_client_access_token().get("access_token"))
        if self.project_field_name in kwargs:
            return super().initial(request, *args, **kwargs)

        for field_name in self.available_project_field_names:
            self.refine_project_field(request.user.token.access_token, field_name, kwargs)

        return super().initial(request, *args, **kwargs)

    def refine_project_field(self, access_token, field_name, kwargs):
        if field_name not in kwargs:
            return

        result = paas_cc.get_project(access_token, kwargs[field_name])
        if result.get("code") != ErrorCode.NoError:
            msg = _("项目Code或者ID不正确: {}").format(result.get("message", ""))
            raise error_codes.APIError.f(msg, replace=True)

        if self.project_field_name == "project_code":
            field_value = result["data"]["english_name"]
        elif self.project_field_name == "project_id":
            field_value = result["data"]["project_id"]
        else:
            field_value = kwargs[field_name]

        kwargs[self.project_field_name] = field_value
        self.kwargs[self.project_field_name] = field_value

        kwargs.pop(field_name, "")
        self.kwargs.pop(field_name, "")

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
from django.conf import settings

from backend.components import paas_cc
from backend.container_service.projects.authorized import list_auth_projects
from backend.container_service.projects.base.constants import ProjectKindName
from backend.uniapps.apis.constants import PAAS_CD_APIGW_PUBLIC_KEY
from backend.utils.authentication import JWTClient
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes
from backend.utils.func_controller import get_func_controller

DEFAULT_APP_CODE = "workbench"
APP_CODE_SKIP_AUTH_WHITE_LIST = "APP_CODE_SKIP_AUTH"


def skip_authentication(app_code):
    """检查app是否在白名单中"""
    # 当功能开关为白名单时，注意下面的含义
    # enable: True/False; True表示此功能完全开放，False表示此功能只针对白名单中的开放
    enabled, wlist = get_func_controller(APP_CODE_SKIP_AUTH_WHITE_LIST)
    if enabled or app_code in wlist:
        return True
    return False


def check_user_project(access_token, project_id, cc_app_id, jwt_info, project_code_flag=False, is_orgin_project=False):
    """检测用户有项目权限"""
    # 针对非用户态进行特殊处理
    if not jwt_info and settings.DEBUG:
        app_code = DEFAULT_APP_CODE
    else:
        app_code, _ = parse_jwt_info(jwt_info)
    if skip_authentication(app_code):
        resp = paas_cc.get_project(access_token, project_id)
        if resp.get("code") != ErrorCode.NoError:
            raise error_codes.APIError.f(resp.get("message", ""))
        data = resp.get("data") or {}
        if str(data.get("cc_app_id")) != str(cc_app_id):
            raise error_codes.CheckFailed.f("用户没有访问项目的权限，请确认", replace=True)
        if project_code_flag:
            project_code = data.get("english_name")
        project_info = data
    else:
        project_info = list_auth_projects(access_token)
        # 通过cc app id过滤
        if project_info.get("code") != ErrorCode.NoError:
            raise error_codes.APIError.f(project_info.get("message"))
        ret_data = [
            info
            for info in project_info.get("data") or []
            if str(info.get("cc_app_id")) == str(cc_app_id) and info["project_id"] == project_id
        ]
        if not ret_data:
            raise error_codes.CheckFailed.f("用户没有访问项目的权限，请确认", replace=True)
        project_info = ret_data[0]
        project_code = project_info.get("english_name")
    # 直接返回原生的project信息
    project_kind = ProjectKindName
    if is_orgin_project:
        return project_kind, app_code, project_info
    if project_code_flag:
        return project_kind, app_code, project_code
    return project_kind, app_code


def parse_jwt_info(jwt_info):
    """解析JWT获取应用/用户/项目等信息"""
    client = JWTClient(jwt_info)
    if not client.is_valid(PAAS_CD_APIGW_PUBLIC_KEY):
        raise error_codes.CheckFailed.f("解析JWT异常，已通知管理员", replace=True)
    app_code = client.app.app_code
    username = client.user.username
    return app_code, username

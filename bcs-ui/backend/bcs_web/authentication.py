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
from rest_framework import exceptions
from rest_framework.authentication import BaseAuthentication

from backend.components.paas_auth import get_access_token
from backend.utils import FancyDict
from backend.utils.authentication import JWTClient, JWTUser
from backend.utils.whitelist import is_app_open_api_trusted

from . import constants


class JWTAndTokenAuthentication(BaseAuthentication):
    """提供用户认证功能(用于接入apigw的网关API)"""

    def authenticate(self, request):
        user = self.authenticate_jwt(request)
        return self.authenticate_token(request, user)

    def authenticate_jwt(self, request) -> JWTUser:
        client = JWTClient(request.META.get(constants.APIGW_JWT_KEY_NAME, ""))
        if not client.is_valid(constants.BCS_APP_APIGW_PUBLIC_KEY):
            raise exceptions.AuthenticationFailed(f"invalid {constants.APIGW_JWT_KEY_NAME}")

        username = client.user.username
        if not username and is_app_open_api_trusted(client.app.app_code):
            username = request.META.get(constants.USERNAME_KEY_NAME, "")

        user = JWTUser(username=username)
        user.client = client
        return user

    def authenticate_token(self, request, user: JWTUser):
        """生成有效的request.user.token.access_token"""

        access_token = request.META.get(constants.ACCESS_TOKEN_KEY_NAME, "")
        # 通过头部传入access_token
        if access_token:
            self._validate_access_token(request, access_token)
            user.token = FancyDict(access_token=access_token)
        else:  # 如果客户端未传入有效access_token, 平台注入系统access_token
            user.token = FancyDict(access_token=get_access_token().get("access_token"))

        return (user, None)

    def _validate_access_token(self, request, access_token):
        # 代码多版本原因: 如果paas_auth中定义了get_user_by_access_token方法，则完成用户身份校验；否则忽略
        try:
            from backend.components.paas_auth import get_user_by_access_token
        except ImportError:
            pass
        else:
            user = get_user_by_access_token(access_token)
            if user.get("user_id") != request.user.username:
                raise exceptions.AuthenticationFailed(f"invalid {constants.ACCESS_TOKEN_KEY_NAME}")

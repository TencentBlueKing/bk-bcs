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

from backend.components import utils
from backend.utils.decorators import parse_response_data

HEADERS = {"X-BK-APP-CODE": settings.APP_ID, "X-BK-APP-SECRET": settings.APP_TOKEN}


@parse_response_data(default_data={})
def get_access_token(params):
    url = f"{settings.BK_SSM_HOST}/api/v1/auth/access-tokens"
    return utils.http_post(url, json=params, headers=HEADERS)


def get_bk_login_access_token(bk_token):
    """获取access_token"""
    return get_access_token({"grant_type": "authorization_code", "id_provider": "bk_login", "bk_token": bk_token})


def get_client_access_token():
    """获取非用户态access_token"""
    return get_access_token({"grant_type": "client_credentials", "id_provider": "client"})


@parse_response_data(default_data={})
def get_authorization_by_access_token(access_token):
    url = f"{settings.BK_SSM_HOST}/api/v1/auth/access-tokens/verify"
    return utils.http_post(url, json={"access_token": access_token}, headers=HEADERS)

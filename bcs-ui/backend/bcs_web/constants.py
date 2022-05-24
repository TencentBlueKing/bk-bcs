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
import base64

from django.conf import settings

ACCESS_TOKEN_KEY_NAME = 'HTTP_X_BKAPI_TOKEN'
APIGW_JWT_KEY_NAME = 'HTTP_X_BKAPI_JWT'
USERNAME_KEY_NAME = 'HTTP_X_BKAPI_USERNAME'


# 获取 bcs-app 网关的 public key
def get_bcs_app_public_key() -> str:
    public_key = getattr(settings, 'BCS_APP_APIGW_PUBLIC_KEY')
    # CE版本获取到的为 base64 编码后的内容，先进行 base64 解码
    if public_key and settings.EDITION == settings.COMMUNITY_EDITION:
        public_key = base64.b64decode(public_key).decode("utf-8")

    return public_key


BCS_APP_APIGW_PUBLIC_KEY = get_bcs_app_public_key()

# 受信任的app可以从header获取用户名.(私有化版本apigw不支持bk_username传参)
trusted_app_list = ["bk_bcs_monitor", "bk_harbor", "bk_bcs", "workbench"]

# 缓存项目信息的标识
bcs_project_cache_key = f"BK_DEVOPS_BCS:ENABLED_BCS_PROJECT:{{project_id_or_code}}"

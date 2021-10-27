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
import os

from django.conf import settings

from backend.components.apigw import get_api_public_key

ACCESS_TOKEN_KEY_NAME = 'HTTP_X_BKAPI_TOKEN'
APIGW_JWT_KEY_NAME = 'HTTP_X_BKAPI_JWT'
USERNAME_KEY_NAME = 'HTTP_X_BKAPI_USERNAME'

try:
    BCS_APP_APIGW_PUBLIC_KEY = getattr(settings, 'BCS_APP_APIGW_PUBLIC_KEY')
except AttributeError:
    BCS_APP_APIGW_PUBLIC_KEY = get_api_public_key('bcs-app', 'bk_bcs', os.environ.get('BKAPP_BK_BCS_TOKEN'))

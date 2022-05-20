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

from django.conf import settings

from backend.utils import requests

logger = logging.getLogger(__name__)


def get_api_public_key(api_name, app_code=None, app_secret=None):
    try:
        url = f"{settings.APIGW_HOST}/apigw/managementapi/get_api_public_key/"
        headers = {"BK-APP-CODE": app_code or settings.APP_ID, "BK-APP-SECRET": app_secret or settings.APP_TOKEN}
        params = {"api_name": api_name}
        data = requests.bk_get(url, params=params, headers=headers)
        return data.get("public_key")
    except Exception:
        logger.error("get api(%s) public key failed", api_name)
        return None

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
from typing import List

from django.conf import settings
from iam import IAM
from iam.apply import models

from .request import ActionResourcesRequest

logger = logging.getLogger(__name__)


class ApplyURLGenerator:
    iam = IAM(
        settings.APP_CODE,
        settings.SECRET_KEY,
        settings.BK_IAM_HOST,
        settings.BK_PAAS_INNER_HOST,
        settings.BK_IAM_APIGATEWAY_URL,
    )

    @classmethod
    def generate_apply_url(cls, username: str, action_request_list: List[ActionResourcesRequest]) -> str:
        """
        生成权限申请跳转 url
        参考 https://github.com/TencentBlueKing/iam-python-sdk/blob/master/docs/usage.md#14-获取无权限申请跳转url
        """
        app = cls._make_application(action_request_list)
        ok, message, url = cls.iam.get_apply_url(app, bk_username=username)
        if not ok:
            logger.error('generate_apply_url failed: %s', message)
            return settings.BK_IAM_APP_URL
        return url

    @staticmethod
    def _make_application(action_request_list: List[ActionResourcesRequest]) -> models.Application:
        """为 generate_apply_url 方法生成 models.Application"""
        return models.Application(settings.BK_IAM_SYSTEM_ID, actions=[req.to_action() for req in action_request_list])

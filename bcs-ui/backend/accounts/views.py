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

from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.response import Response
from rest_framework.views import APIView

from backend.accounts import bcs_perm
from backend.utils.renderers import BKAPIRenderer

logger = logging.getLogger(__name__)


class UserInfoViewSet(APIView):
    """
    用户信息相关
    """

    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get(self, request):
        user = request.user

        data = {
            "avatar_url": "",
            "username": user.username,
        }
        return Response(data)

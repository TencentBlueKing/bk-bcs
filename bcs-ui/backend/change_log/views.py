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

from rest_framework import permissions
from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.response import Response

from backend.bcs_web.viewsets import SystemViewSet
from backend.utils.renderers import BKAPIRenderer

from .change_log import ChangeLog

logger = logging.getLogger(__name__)


class ChangeLogViewSet(SystemViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)
    permission_classes = (permissions.IsAuthenticated,)

    def list(self, request):
        """展示markdown格式的版本列表"""
        try:
            change_logs = ChangeLog(language=request.LANGUAGE_CODE).list()
        except Exception as e:
            # 当解析异常时，仅记录日志，不影响服务
            logger.exception("获取changelog失败, %s", e)
            change_logs = []
        return Response(change_logs)

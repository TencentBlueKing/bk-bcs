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
import re
from urllib import parse

from channels.auth import AuthMiddlewareStack
from django.http.response import HttpResponseForbidden
from django.utils.deprecation import MiddlewareMixin
from django.utils.translation import ugettext_lazy as _

from backend.container_service.clusters.base.models import CtxCluster
from backend.utils.local import local
from backend.web_console.session import session_mgr

logger = logging.getLogger(__name__)


class DisableCSRFCheck(MiddlewareMixin):
    """本地开发，去掉django rest framework强制的csrf检查"""

    def process_request(self, request):
        setattr(request, '_dont_enforce_csrf_checks', True)


class RequestProvider(object):
    """request_id中间件
    调用链使用
    """

    def __init__(self, get_response=None):
        self.get_response = get_response

    def __call__(self, request):
        local.request = request
        request.request_id = local.get_http_request_id()

        response = self.get_response(request)
        response['X-Request-Id'] = request.request_id

        local.release()

        return response

    # Compatibility methods for Django <1.10
    def process_request(self, request):
        local.request = request
        request.request_id = local.get_http_request_id()

    def process_response(self, request, response):
        response['X-Request-Id'] = request.request_id
        local.release()
        return response


class ChannelSessionAuthMiddleware:
    """django channel auth middleware"""

    CHANNEL_URL_PATTERN = re.compile(r'/projects/(?P<project_id>\w{32})/clusters/(?P<cluster_id>[\w\-]+)/')

    def __init__(self, inner):
        self.inner = inner

    def extract_project_and_cluster_id(self, scope):
        pattern = re.findall(self.CHANNEL_URL_PATTERN, scope['path'])
        if len(pattern) == 0:
            raise HttpResponseForbidden(_("url中project_id或者cluster_id为空"))

        return pattern[0]

    async def __call__(self, scope, receive, send):
        query_params = dict(parse.parse_qsl(scope['query_string'].decode('utf8')))

        session_id = query_params.get("session_id", None)
        if not session_id:
            raise HttpResponseForbidden(_("session_id为空"))

        project_id, cluster_id = self.extract_project_and_cluster_id(scope)

        session = session_mgr.create(project_id, cluster_id)
        ctx = session.get(session_id)
        if not ctx:
            raise HttpResponseForbidden(_("获取ctx为空, session_id不正确或者已经过期"))

        ctx_cluster = CtxCluster.create(
            id=ctx['cluster_id'],
            project_id=ctx['project_id'],
            token=ctx['access_token'],
        )

        scope["ctx_cluster"] = ctx_cluster
        scope["ctx_session"] = ctx

        return await self.inner(scope, receive, send)


# Handy shortcut for applying all three layers at once
def BCSChannelAuthMiddlewareStack(inner):
    return ChannelSessionAuthMiddleware(AuthMiddlewareStack(inner))


try:
    from .middlewares_ext import *  # noqa
except ImportError as e:
    logger.debug('Load extension failed: %s', e)

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
from rest_framework import permissions, response
from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.views import APIView

from backend.components.proxy import ProxyClient, ProxyConfig
from backend.utils.renderers import BKAPIRenderer


class ClusterManagerProxyViewSet(APIView):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)
    permission_classes = ()

    def get(self, request, *args, **kwargs):
        return self._request(request)

    def put(self, request, *args, **kwargs):
        return self._request(request)

    def post(self, request, *args, **kwargs):
        return self._request(request)

    def delete(self, request, *args, **kwargs):
        return self._request(request)

    def _request(self, request) -> response.Response:
        proxy_data = self._get_proxy_data(request)
        return response.Response(ProxyClient(proxy_data).proxy())

    def _get_proxy_data(self, request) -> ProxyConfig:
        proxy_config = settings.CLUSTER_MANAGER_PROXY
        return ProxyConfig(
            host=proxy_config["HOST"],
            request=request,
            prefix_path=proxy_config["PREFIX_PATH"],
            token=proxy_config["TOKEN"],
        )

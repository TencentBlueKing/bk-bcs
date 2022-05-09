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
import copy
from typing import Any, Dict, Optional

from rest_framework import permissions, viewsets
from rest_framework.renderers import BrowsableAPIRenderer

from backend.utils.renderers import BKAPIRenderer

from .authentication import JWTAndTokenAuthentication
from .permissions import AccessProjectPermission, ProjectEnableBCS


class GenericMixin:
    @staticmethod
    def get_request_data(request, **kwargs) -> Dict[str, Any]:
        request_data = request.data.copy() or {}
        request_data.update(**kwargs)
        return request_data

    def params_validate(
        self, serializer, context: Optional[Dict] = None, init_params: Optional[Dict] = None, **kwargs
    ):
        """
        检查参数是够符合序列化器定义的通用逻辑

        :param serializer: 序列化器
        :param context: 上下文数据，可在 View 层传入给 SLZ 使用
        :param init_params: 初始参数，替换掉 request.query_params / request.data
        :param kwargs: 可变参数
        :return: 校验的结果
        """
        if init_params is None:
            # 获取 Django request 对象
            _request = self.request
            if _request.method in ['GET', 'DELETE']:
                req_data = copy.deepcopy(_request.query_params)
            else:
                req_data = _request.data.copy()

            # NOTE 兼容措施，确保未切换的老接口 Delete 请求还可以从 request.body 中获取请求参数
            # 新接口 Delete 请求参数应从 query_params 中获取，复杂参数如批量删除的 list，该用 Post 请求
            if _request.method == 'DELETE' and not req_data:
                req_data = _request.data.copy()
        else:
            req_data = init_params

        if kwargs:
            req_data.update(kwargs)

        # 参数校验，如不符合直接抛出异常
        context = context or {}
        slz = serializer(data=req_data, context=context)
        slz.is_valid(raise_exception=True)
        return slz.validated_data


class SystemViewSet(GenericMixin, viewsets.ViewSet):
    """
    容器服务 SaaS app 使用的 API view
    - 仅支持处理 url 路径参数中包含 project_id 或 project_id_or_code 的请求
    - 需要验证用户登录态
    - ProjectEnableBCS: 验证通过后，会创建 request.project、request.ctx_project 和 request.ctx_cluster，在 view 中使用
    """

    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)
    # authentication_classes配置在REST_FRAMEWORK["DEFAULT_AUTHENTICATION_CLASSES"]中
    permission_classes = (permissions.IsAuthenticated, AccessProjectPermission, ProjectEnableBCS)


class UserViewSet(GenericMixin, viewsets.ViewSet):
    """
    提供给流水线等第三方服务的API view
    - 仅支持处理 url 路径参数中包含 project_id 或 project_id_or_code 的请求
    - JWTAndTokenAuthentication: 需要传入有效的 JWT 以便验证请求来源; 同时，也会将 HTTP_X_BKAPI_TOKEN 或平台的 access_token
    赋值给 request.user.token.access_token
    - ProjectEnableBCS: 验证通过后，会创建 request.project、request.ctx_project 和 request.ctx_cluster，在 view 中使用
    """

    renderer_classes = (BKAPIRenderer,)
    authentication_classes = (JWTAndTokenAuthentication,)
    permission_classes = (permissions.IsAuthenticated, AccessProjectPermission, ProjectEnableBCS)

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
from typing import Dict, Optional

from rest_framework import viewsets
from rest_framework.authentication import BaseAuthentication
from rest_framework.permissions import BasePermission
from rest_framework.renderers import BrowsableAPIRenderer

from backend.utils import FancyDict
from backend.utils.renderers import BKAPIRenderer


class FakeUserAuth(BaseAuthentication):
    """ 假的用户身份认证类，单元测试用 """

    def authenticate(self, request):
        class APIUserToken:
            access_token = 'fake_access_token'

        class APIUser:
            token = APIUserToken
            is_superuser = False
            username = 'user_for_test'

        if not hasattr(request, '_user'):
            request._user = APIUser
        elif not hasattr(request._user, 'token'):
            request._user.token = APIUserToken

        return (APIUser, APIUserToken)


class FakeProjectEnableBCS(BasePermission):
    """ 假的权限控制类，单元测试用 """

    def has_permission(self, request, view):
        project_id = view.kwargs.get('project_id', '') or view.kwargs.get('project_id_or_code', '')
        # project 内容为 StubPaaSCCClient 所 mock，如需要自定义 project，可以参考
        # backend/tests/container_service/clusters/open_apis/test_namespace.py:53
        request.project = self._get_enabled_project('fake_access_token', project_id)
        self._set_ctx_project_cluster(request, project_id, view.kwargs.get('cluster_id', ''))
        return True

    def _get_enabled_project(self, access_token, project_id_or_code: str) -> Optional[FancyDict]:
        from backend.tests.testing_utils.mocks.paas_cc import StubPaaSCCClient

        project_data = StubPaaSCCClient().get_project(project_id_or_code)
        project = FancyDict(**project_data)

        if project.cc_app_id != 0:
            return project

        return None

    def _set_ctx_project_cluster(self, request, project_id: str, cluster_id: str):
        from backend.container_service.clusters.base.models import CtxCluster
        from backend.container_service.projects.base.models import CtxProject

        access_token = 'access_token_for_test'
        request.ctx_project = CtxProject.create(token=access_token, id=project_id)
        if cluster_id:
            request.ctx_cluster = CtxCluster.create(token=access_token, id=cluster_id, project_id=project_id)
        else:
            request.ctx_cluster = None


class SimpleGenericMixin:
    """
    backend.bcs_web.viewsets.GenericMixin 精简版
    根据实际需要，挪相关方法用于单元测试 Mock
    """

    def params_validate(
        self, serializer, context: Optional[Dict] = None, init_params: Optional[Dict] = None, **kwargs
    ):
        """
        检查参数是够符合序列化器定义的通用逻辑

        :param serializer: 序列化器
        :param init_params: 初始参数
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
        context = context if context else {}
        slz = serializer(data=req_data, context=context)
        slz.is_valid(raise_exception=True)
        return slz.validated_data


class FakeSystemViewSet(SimpleGenericMixin, viewsets.ViewSet):
    """ 假的基类 ViewSet，单元测试用 """

    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)
    # 替换掉原来的 认证 / 权限控制类
    authentication_classes = (FakeUserAuth,)
    permission_classes = (FakeProjectEnableBCS,)


class FakeUserViewSet(FakeSystemViewSet):
    """ 假的用户基类 ViewSet，单元测试用 """

    renderer_classes = (BKAPIRenderer,)
    authentication_classes = tuple()

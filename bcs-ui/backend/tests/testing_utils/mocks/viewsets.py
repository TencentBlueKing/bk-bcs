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
from rest_framework.permissions import BasePermission
from rest_framework.renderers import BrowsableAPIRenderer

from backend.utils import FancyDict
from backend.utils.renderers import BKAPIRenderer


class FakeProjectEnableBCS(BasePermission):
    """ 假的权限控制类，单元测试用 """

    def has_permission(self, request, view):
        project_id = view.kwargs.get('project_id', '')
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

    def params_validate(self, serializer, init_params: Optional[Dict] = None, **kwargs):
        """
        检查参数是够符合序列化器定义的通用逻辑

        :param serializer: 序列化器
        :param init_params: 初始参数
        :param kwargs: 可变参数
        :return: 校验的结果
        """
        if init_params is None:
            # 获取 Django request 对象
            _request = self.request
            if _request.method in ['GET']:
                req_data = copy.deepcopy(_request.query_params)
            else:
                req_data = _request.data.copy()
        else:
            req_data = init_params

        if kwargs:
            req_data.update(kwargs)

        # 参数校验，如不符合直接抛出异常
        slz = serializer(data=req_data)
        slz.is_valid(raise_exception=True)
        return slz.validated_data


class FakeSystemViewSet(SimpleGenericMixin, viewsets.ViewSet):
    """ 假的基类 ViewSet，单元测试用 """

    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)
    # 替换掉原有的权限控制类
    permission_classes = (FakeProjectEnableBCS,)


class FakeUserViewSet(FakeSystemViewSet):
    """ 假的用户基类 ViewSet，单元测试用 """

    renderer_classes = (BKAPIRenderer,)
    authentication_classes = tuple()

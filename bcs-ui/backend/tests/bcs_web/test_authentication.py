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
import mock
import pytest
from django.conf.urls import url
from rest_framework import permissions, status
from rest_framework.response import Response
from rest_framework.test import RequestsClient
from rest_framework.views import APIView

from backend.bcs_web.authentication import JWTAndTokenAuthentication
from backend.tests.testing_utils.mocks import jwt

pytestmark = pytest.mark.django_db


class MockView(APIView):
    permission_classes = (permissions.IsAuthenticated,)

    def get(self, request):
        return Response({'access_token': request.user.token.access_token})


@pytest.fixture(autouse=True)
def patch_authentication():
    with mock.patch('backend.bcs_web.authentication.JWTClient', new=jwt.FakeJWTClient), mock.patch(
        'backend.bcs_web.authentication.JWTAndTokenAuthentication._validate_access_token',
        new=lambda *args, **kwargs: True,
    ):
        yield


urlpatterns = [
    url('test/', MockView.as_view(authentication_classes=[JWTAndTokenAuthentication])),
]


@pytest.mark.urls(__name__)
class TestJWTAndTokenAuthentication:
    def test_authentication(self):
        client = RequestsClient()

        # 未提供用户认证信息
        response = client.get('http://testserver/test/')
        assert response.status_code == status.HTTP_401_UNAUTHORIZED

        # 提供用户认证信息
        access_token = 'test_access_token'
        response = client.get(
            'http://testserver/test/', headers={'X-BKAPI-JWT': jwt.VALID_JWT, 'X-BKAPI-TOKEN': access_token}
        )
        assert response.status_code == status.HTTP_200_OK
        assert response.json()['access_token'] == access_token

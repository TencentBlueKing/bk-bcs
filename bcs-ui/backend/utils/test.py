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
from django.contrib.auth import get_user_model
from rest_framework import test

User = get_user_model()


def get_testing_user(**kwargs):
    user = User.objects.create(
        username="testing",
    )
    for k, v in kwargs:
        setattr(user, k, v)
    user.save()
    user.token = mock.MagicMock()
    user.token.access_token = "testing"
    user.token.expires_soon = lambda: False
    return user


class APITestCase(test.APITransactionTestCase):
    def get_default_user(self):
        user = getattr(self, "user", None)
        if user:
            return user
        self.user = get_testing_user()
        return self.user

    def force_authenticate(self, user=None):
        user = user or self.get_default_user()
        self.client.force_authenticate(self.user)  # noqa

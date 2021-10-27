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
import rest_framework.authentication
import rest_framework.exceptions
from django.utils.translation import ugettext_lazy as _

from backend.utils import FancyDict

from .models import Token


class TokenAuthentication(rest_framework.authentication.TokenAuthentication):
    model = Token

    def authenticate(self, request):
        auth = rest_framework.authentication.get_authorization_header(request).split()

        if not auth or auth[0].lower() != self.keyword.lower().encode():
            return None

        if len(auth) == 1:
            msg = _('Invalid token header. No credentials provided.')
            raise rest_framework.exceptions.AuthenticationFailed(msg)
        elif len(auth) > 2:
            msg = _('Invalid token header. Token string should not contain spaces.')
            raise rest_framework.exceptions.AuthenticationFailed(msg)

        try:
            token = auth[1].decode()
        except UnicodeError:
            msg = _('Invalid token header. Token string should not contain invalid characters.')
            raise rest_framework.exceptions.AuthenticationFailed(msg)

        model = self.get_model()
        try:
            token = model.objects.get(key=token)
        except model.DoesNotExist:
            msg = _('Invalid token header. Token string not reisted.')
            raise rest_framework.exceptions.AuthenticationFailed(msg)

        user = FancyDict(
            token=token,
            is_authenticated=True,
            username=token.username,
        )
        return user, token

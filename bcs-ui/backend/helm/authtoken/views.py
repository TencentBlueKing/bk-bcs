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

from rest_framework import viewsets

from backend.utils.views import ActionSerializerMixin, with_code_wrapper

from .models import Token
from .serializers import TokenSLZ, TokenUpdateSLZ

logger = logging.getLogger(__name__)


@with_code_wrapper
class TokenView(ActionSerializerMixin, viewsets.ModelViewSet):
    serializer_class = TokenSLZ
    lookup_url_kwarg = "token_id"

    action_serializers = {
        'update': TokenUpdateSLZ,
    }

    def get_queryset(self):
        return Token.objects.filter(username=self.request.user.username)

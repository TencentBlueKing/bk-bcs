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
import jsonschema
from django.db import models
from rest_framework.exceptions import PermissionDenied

from .exceptions import ObjectAreadyExist
from .providers import provider_map


class TokenManager(models.Manager):
    def load_token_provider_cls(self, kind):
        return provider_map[kind]

    def make_token(self, user, name, kind, config, description, maintainers):
        if self.filter(username=user.username, name=name).exists():
            raise ObjectAreadyExist

        provider_cls = self.load_token_provider_cls(kind)
        # TODO deal with ValidationError
        jsonschema.validate(config, provider_cls.CONFIG_SCHEMA)

        # TODO deal with PermissionDenied
        configuration = provider_cls.provide(user, config)

        token = self.create(
            name=name,
            kind=kind,
            username=user.username,
            config=configuration,
            description=description,
            maintainers=maintainers,
        )
        return token

    def validate_request_data(self, token, request_data):
        provider_cls = self.load_token_provider_cls(token.kind)

        jsonschema.validate(request_data, provider_cls.REQUEST_SCHEMA)

        provider_cls.validate(token, request_data)

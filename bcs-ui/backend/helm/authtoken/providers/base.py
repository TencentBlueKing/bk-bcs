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
from rest_framework.exceptions import PermissionDenied


class BaseProvider:
    """base token kind provider"""

    NAME = ""
    CONFIG_SCHEMA = {}
    REQUEST_SCHEMA = {}

    @staticmethod
    def provide(user, project_id, config):
        """this method validate whether the user could provide this kind of token with config.
        if user doesn't have the priority, this method must raise a `rest_framework.exceptions.PermissionDenied`
        ex: config contains helm app id, provider must validate user can operate it.
        params user: user of UnionAuth type
        params config: data matchs SCHEMA, it contains necessary data for apply this kind of token.
        return: it must return a which can be serialize to json
        """
        raise NotImplementedError

    @staticmethod
    def validate(token, request_data):
        """
        justice whether user specified by token could do the operation.
        if user doesn't have the priority, this method must raise a `rest_framework.exceptions.PermissionDenied`
        params token: object of .models.Token which indicate request user
        params request_data: parameters for do operation
        """
        raise NotImplementedError

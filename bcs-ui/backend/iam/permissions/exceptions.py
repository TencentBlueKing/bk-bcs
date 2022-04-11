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
from typing import Dict, List, Optional

from .apply_url import ApplyURLGenerator
from .request import ActionResourcesRequest


class PermissionDeniedError(Exception):
    message = 'Permission denied'
    code = 40300

    def __init__(
        self,
        message: str = '',
        username: str = '',
        action_request_list: Optional[List[ActionResourcesRequest]] = None,
    ):
        """
        :param message: 异常信息
        :param username: 无权限的用户名, 用于生成 apply_url
        :param action_request_list: 用于生成 apply_url
        """

        if message:
            self.message = message

        self.username = username
        self.action_request_list = action_request_list

    @property
    def data(self) -> Dict:
        return {
            'perms': {
                'apply_url': ApplyURLGenerator.generate_apply_url(self.username, self.action_request_list),
                'action_list': [
                    {'resource_type': action_request.resource_type, 'action_id': action_request.action_id}
                    for action_request in self.action_request_list
                ],
            }
        }

    def __str__(self):
        return f'{self.code}: {self.message}'


class AttrValidationError(Exception):
    """属性字段校验异常"""

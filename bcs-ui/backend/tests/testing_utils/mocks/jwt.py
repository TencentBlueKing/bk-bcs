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
from backend.utils import FancyDict

VALID_JWT = "Header.Payload.Signature"


class FakeJWTClient:
    def __init__(self, content):
        self.content = content
        self.payload = {}
        self.headers = {}

    @property
    def user(self):
        return FancyDict(self.payload.get('user') or {})

    @property
    def app(self):
        return FancyDict(self.payload.get('app') or {})

    def is_valid(self, apigw_public_key=None):
        if self.content == VALID_JWT:
            self.payload = {'user': {'username': 'bcs_admin'}, 'app': {'app_code': 'bcs-app'}}
            return True
        else:
            return False

    def __str__(self):
        return '<%s, %s>' % (self.headers, self.payload)

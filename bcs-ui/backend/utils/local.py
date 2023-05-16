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

全局相关
"""
import uuid

from django.conf import settings
from django.http import HttpRequest
from werkzeug.local import Local as _Local
from werkzeug.local import release_local

from backend.utils import FancyDict

_local = _Local()


def new_request_id():
    return uuid.uuid4().hex


class Singleton(object):
    _instance = None

    def __new__(cls, *args, **kwargs):
        if not isinstance(cls._instance, cls):
            cls._instance = object.__new__(cls, *args, **kwargs)
        return cls._instance


class Local(Singleton):
    """local对象
    必须配合中间件RequestProvider使用
    """

    @property
    def request(self):
        """获取全局request对象"""
        request = getattr(_local, 'request', None)
        # if not request:
        #     raise RuntimeError("request object not in local")
        return request

    @request.setter
    def request(self, value):
        """设置全局request对象"""
        _local.request = value

    def new_dummy_request(self, access_token, username):
        """celery 后台任务等主动设置"""
        if self.request is not None:
            return

        request = HttpRequest()
        token = FancyDict({"access_token": access_token})
        request.user = FancyDict({"token": token, "username": username})
        request.request_id = new_request_id()
        _local.request = request

    @property
    def request_id(self):
        # celery后台没有request对象
        if self.request:
            return self.request.request_id

        return new_request_id()

    def get_http_request_id(self):
        """从接入层获取request_id，或者生成一个新的request_id"""
        # 在从header中获取
        request_id = self.request.META.get(settings.REQUEST_ID_HEADER, '')
        if request_id:
            return request_id

        # 最后主动生成一个
        return new_request_id()

    def release(self):
        release_local(_local)


local = Local()

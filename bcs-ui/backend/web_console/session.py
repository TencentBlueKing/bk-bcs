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
import abc
import copy
import json
import logging
import uuid

from backend.utils.cache import rd_client
from backend.utils.renderers import JSONEncoder

logger = logging.getLogger(__name__)

DEFAULT_SESSION_TYPE = 'redis'


class SessionBase(abc.ABC):
    CACHE_KEY = 'WebConsole:{project_id}:{cluster_id}:{session_id}'

    # 默认一小时过期
    EXPIRE = 3600

    def __init__(self, project_id, cluster_id):
        self.project_id = project_id
        self.cluster_id = cluster_id

    @abc.abstractmethod
    def set(self, ctx: dict) -> str:
        """保存ctx到session"""

    @abc.abstractmethod
    def get(self, session_id: str) -> dict:
        """通过session_id获取ctx"""


class RedisSession(SessionBase):
    def set(self, ctx: dict) -> str:
        ctx = copy.deepcopy(ctx)

        # 获取web-console context信息
        session_id = uuid.uuid4().hex

        ctx['session_id'] = session_id

        key = self.CACHE_KEY.format(project_id=self.project_id, cluster_id=self.cluster_id, session_id=session_id)

        rd_client.set(key, json.dumps(ctx, cls=JSONEncoder), ex=self.EXPIRE)

        return session_id

    def get(self, session_id):
        key = self.CACHE_KEY.format(project_id=self.project_id, cluster_id=self.cluster_id, session_id=session_id)
        raw_ctx = rd_client.get(key)
        if not raw_ctx:
            logger.info("session_id 不正确或者已经过期, %s", session_id)
            return None

        try:
            ctx = json.loads(raw_ctx)
        except Exception as error:
            logger.info("bcs_context 格式不是json, %s, %s", raw_ctx, error)
            ctx = None

        return ctx


class SessionMgr:
    def __init__(self):
        self._session_cls = {}

    def register(self, _type: str, session_cls):
        self._session_cls[_type] = session_cls

    def create(self, project_id: str, cluster_id: str, _type: str = DEFAULT_SESSION_TYPE):
        session_cls = self._session_cls.get(_type)
        if not session_cls:
            raise ValueError(f'{_type} not in {self._session_cls}')
        return session_cls(project_id, cluster_id)


session_mgr = SessionMgr()
session_mgr.register('redis', RedisSession)

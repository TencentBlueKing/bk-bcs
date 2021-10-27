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
from functools import wraps

import tornado.web
from django.conf import settings
from django.utils.translation import ugettext_lazy as _

from backend.components.utils import http_get

from .session import session_mgr

logger = logging.getLogger(__name__)


def authenticated(view_func):
    """权限认证装饰器"""

    @wraps(view_func)
    def _wrapped_view(self, *args, **kwargs):
        session_id = self.get_argument("session_id", None)
        if not session_id:
            raise tornado.web.HTTPError(401, log_message=_("session_id为空"))

        project_id = kwargs.get("project_id", "")
        cluster_id = kwargs.get("cluster_id", "")
        session = session_mgr.create(project_id, cluster_id)

        ctx = session.get(session_id)
        if not ctx:
            raise tornado.web.HTTPError(401, log_message=_("获取ctx为空, session_id不正确或者已经过期"))

        kwargs["context"] = ctx

        return view_func(self, *args, **kwargs)

    return _wrapped_view

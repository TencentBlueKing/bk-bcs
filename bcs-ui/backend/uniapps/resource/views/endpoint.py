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

from django.utils.translation import ugettext_lazy as _

from backend.components.bcs import k8s
from backend.container_service.clusters.base.utils import get_cluster_type
from backend.container_service.clusters.constants import ClusterType
from backend.uniapps.application.base_views import BaseAPI
from backend.uniapps.application.utils import APIResponse
from backend.utils.errcodes import ErrorCode

logger = logging.getLogger(__name__)


class Endpoints(BaseAPI):
    def get(self, request, project_id, cluster_id, namespace, name):
        """获取项目下所有的endpoints"""
        if get_cluster_type(cluster_id) == ClusterType.SHARED:
            return APIResponse({"code": 400, "message": _("无法查看共享集群资源")})

        params = {"name": name, "namespace": namespace}
        client = k8s.K8SClient(request.user.token.access_token, project_id, cluster_id, env=None)
        resp = client.get_endpoints(params)

        if resp.get("code") != ErrorCode.NoError:
            return APIResponse(
                {"code": resp.get("code", ErrorCode.UnknownError), "message": resp.get("message", _("请求出现异常!"))}
            )

        return APIResponse({"code": ErrorCode.NoError, "data": resp.get("data"), "message": "ok"})

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
from rest_framework.response import Response

from backend.bcs_web.viewsets import UserViewSet
from backend.container_service.clusters.permissions import AccessClusterPermMixin
from backend.resources.utils.format import ResourceDefaultFormatter
from backend.resources.workloads.pod import Pod


class PodViewSet(AccessClusterPermMixin, UserViewSet):
    def get_pod(self, request, project_id_or_code, cluster_id, namespace, pod_name):
        """获取指定 Pod 信息，以列表格式返回"""
        pod = Pod(request.ctx_cluster).get(namespace=namespace, name=pod_name, formatter=ResourceDefaultFormatter())
        # 保持接口格式不变，如果查询不到则返回空列表
        response_data = [pod] if pod else []
        return Response(response_data)

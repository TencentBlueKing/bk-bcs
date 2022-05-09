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
from django.utils.translation import ugettext_lazy as _
from kubernetes.client import ApiException
from rest_framework.response import Response

from backend.bcs_web.viewsets import SystemViewSet
from backend.container_service.clusters.base.utils import get_cluster_type
from backend.dashboard.exceptions import ResourceVersionExpired
from backend.dashboard.subscribe.constants import DEFAULT_SUBSCRIBE_TIMEOUT, K8S_API_GONE_STATUS_CODE
from backend.dashboard.subscribe.permissions import IsSubscribeable
from backend.dashboard.subscribe.serializers import FetchResourceWatchResultSLZ
from backend.dashboard.subscribe.utils import get_native_kind_resource_client, is_native_kind
from backend.resources.constants import K8sResourceKind
from backend.resources.custom_object import CustomObject
from backend.resources.custom_object.formatter import CustomObjectCommonFormatter
from backend.utils.basic import getitems


class SubscribeViewSet(SystemViewSet):
    """订阅相关接口，检查 K8S 资源变更情况"""

    def get_permissions(self):
        return [*super().get_permissions(), IsSubscribeable()]

    def list(self, request, project_id, cluster_id):
        """获取指定资源某resource_version后变更记录"""
        params = self.params_validate(FetchResourceWatchResultSLZ, context={'ctx_cluster': request.ctx_cluster})

        res_kind, res_version, namespace = params['kind'], params['resource_version'], params.get('namespace')
        watch_kwargs = {
            'namespace': namespace,
            'resource_version': res_version,
            'timeout': DEFAULT_SUBSCRIBE_TIMEOUT,
        }
        if is_native_kind(res_kind):
            # 根据 Kind 获取对应的 K8S Resource Client 并初始化
            resource_client = get_native_kind_resource_client(res_kind)(request.ctx_cluster)
            # 对于命名空间，watch_kwargs 需要补充 cluster_type，project_code 以支持共享集群的需求
            if res_kind == K8sResourceKind.Namespace.value:
                watch_kwargs.update(
                    {'cluster_type': get_cluster_type(cluster_id), 'project_code': request.project.english_name}
                )
        else:
            # 自定义资源类型走特殊的获取 ResourceClient 逻辑 且 需要指定 Formatter
            resource_client = CustomObject(request.ctx_cluster, kind=res_kind, api_version=params['api_version'])
            watch_kwargs['formatter'] = CustomObjectCommonFormatter()

        try:
            events = resource_client.watch(**watch_kwargs)
        except ApiException as e:
            if e.status == K8S_API_GONE_STATUS_CODE:
                raise ResourceVersionExpired(_('ResourceVersion {} 已过期，请重新获取').format(res_version))
            raise

        # events 默认按时间排序，取最后一个 ResourceVersion 即为最新值
        latest_rv = getitems(events[-1], 'manifest.metadata.resourceVersion') if events else None
        response_data = {'events': events, 'latest_rv': latest_rv}
        return Response(response_data)

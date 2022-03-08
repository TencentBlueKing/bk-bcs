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
from rest_framework import viewsets
from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.response import Response

from backend.components.bcs import k8s
from backend.container_service.clusters.base.utils import get_cluster_type
from backend.container_service.clusters.constants import ClusterType
from backend.iam.permissions.decorators import response_perms
from backend.iam.permissions.resources.namespace import NamespaceRequest
from backend.iam.permissions.resources.namespace_scoped import NamespaceScopedAction, NamespaceScopedPermission
from backend.templatesets.legacy_apps.configuration.k8s.serializers import K8sIngressSLZ
from backend.templatesets.legacy_apps.instance.constants import K8S_INGRESS_SYS_CONFIG
from backend.uniapps import utils as app_utils
from backend.uniapps.application.base_views import BaseAPI
from backend.uniapps.resource.views import ResourceOperate
from backend.utils.errcodes import ErrorCode
from backend.utils.renderers import BKAPIRenderer
from backend.utils.response import PermsResponse

logger = logging.getLogger(__name__)


class IngressResource(viewsets.ViewSet, BaseAPI, ResourceOperate):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    category = 'ingress'
    cate = 'K8sIngress'
    sys_config = K8S_INGRESS_SYS_CONFIG

    def get_ingress_by_cluser_id(self, request, params, project_id, cluster_id):
        """查询configmaps"""
        if get_cluster_type(cluster_id) == ClusterType.SHARED:
            return 0, []

        access_token = request.user.token.access_token
        client = k8s.K8SClient(access_token, project_id, cluster_id, env=None)
        resp = client.get_ingress(params)

        if resp.get("code") != ErrorCode.NoError:
            logger.error(u"bcs_api error: %s" % resp.get("message", ""))
            return resp.get("code", ErrorCode.UnknownError), resp.get("message", _("请求出现异常!"))
        data = resp.get("data") or []
        return 0, data

    @response_perms(
        action_ids=[NamespaceScopedAction.VIEW, NamespaceScopedAction.UPDATE, NamespaceScopedAction.DELETE],
        permission_cls=NamespaceScopedPermission,
        resource_id_key='iam_ns_id',
    )
    def get(self, request, project_id):
        """获取项目下的所有Ingress"""
        cluster_id = request.query_params.get("cluster_id")

        code, cluster_ingress = self.get_ingress_by_cluser_id(request, {}, project_id, cluster_id)
        # 单个集群错误时，不抛出异常信息
        if code != ErrorCode.NoError:
            return Response({'code': code, 'message': cluster_ingress})

        self.handle_data(
            cluster_ingress,
            self.cate,
            cluster_id,
            False,
            namespace_dict=app_utils.get_ns_id_map(request.user.token.access_token, project_id),
        )

        # 按时间倒序排列
        cluster_ingress.sort(key=lambda x: x.get('createTime', ''), reverse=True)

        return PermsResponse(cluster_ingress, NamespaceRequest(project_id=project_id, cluster_id=cluster_id))

    def delete_ingress(self, request, project_id, cluster_id, namespace, name):
        return self.delete_resource(request, project_id, cluster_id, namespace, name)

    def batch_delete_ingress(self, request, project_id):
        return self.batch_delete_resource(request, project_id)

    def update_ingress(self, request, project_id, cluster_id, namespace, name):
        self.slz = K8sIngressSLZ
        return self.update_resource(request, project_id, cluster_id, namespace, name)

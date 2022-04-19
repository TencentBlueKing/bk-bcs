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
import copy
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
from backend.templatesets.legacy_apps.configuration.serializers import K8sSecretCreateOrUpdateSLZ
from backend.templatesets.legacy_apps.instance.constants import K8S_SECRET_SYS_CONFIG
from backend.uniapps import utils as app_utils
from backend.uniapps.application.base_views import BaseAPI
from backend.uniapps.resource.constants import DEFAULT_SEARCH_FIELDS
from backend.utils.basic import str2bool
from backend.utils.errcodes import ErrorCode
from backend.utils.renderers import BKAPIRenderer
from backend.utils.response import PermsResponse

from .base import ResourceOperate

logger = logging.getLogger(__name__)


class Secrets(viewsets.ViewSet, BaseAPI, ResourceOperate):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    category = 'secret'
    cate = 'K8sSecret'
    sys_config = K8S_SECRET_SYS_CONFIG

    def get_secrets_by_cluster_id(self, request, params, project_id, cluster_id):
        """查询secrets"""
        if get_cluster_type(cluster_id) == ClusterType.SHARED:
            return 0, []

        search_fields = copy.deepcopy(DEFAULT_SEARCH_FIELDS)

        search_fields.append("data.data")
        params.update({"field": ",".join(search_fields)})
        client = k8s.K8SClient(request.user.token.access_token, project_id, cluster_id, env=None)
        resp = client.get_secret(params)

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
        """ 获取项目下所有的secrets """
        params = dict(request.GET.items())
        is_decode = str2bool(request.GET.get('decode'))

        cluster_id = params['cluster_id']

        code, cluster_secrets = self.get_secrets_by_cluster_id(request, params, project_id, cluster_id)
        if code != ErrorCode.NoError:
            return Response({'code': code, 'message': cluster_secrets})

        self.handle_data(
            cluster_secrets,
            self.cate,
            cluster_id,
            is_decode,
            namespace_dict=app_utils.get_ns_id_map(request.user.token.access_token, project_id),
        )

        # 按时间倒序排列
        cluster_secrets.sort(key=lambda x: x.get('createTime', ''), reverse=True)

        return PermsResponse(
            cluster_secrets,
            NamespaceRequest(project_id=project_id, cluster_id=cluster_id),
        )

    def delete_secret(self, request, project_id, cluster_id, namespace, name):
        return self.delete_resource(request, project_id, cluster_id, namespace, name)

    def batch_delete_secrets(self, request, project_id):
        return self.batch_delete_resource(request, project_id)

    def update_secret(self, request, project_id, cluster_id, namespace, name):
        self.slz = K8sSecretCreateOrUpdateSLZ
        return self.update_resource(request, project_id, cluster_id, namespace, name)

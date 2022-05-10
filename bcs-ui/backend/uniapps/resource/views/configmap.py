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

from backend.components import paas_cc
from backend.components.bcs import k8s
from backend.container_service.clusters.base.utils import get_cluster_type
from backend.container_service.clusters.constants import ClusterType
from backend.iam.permissions.decorators import response_perms
from backend.iam.permissions.resources.namespace import NamespaceRequest
from backend.iam.permissions.resources.namespace_scoped import NamespaceScopedAction, NamespaceScopedPermission
from backend.resources.namespace.constants import K8S_SYS_NAMESPACE
from backend.templatesets.legacy_apps.configuration.serializers import K8sConfigMapCreateOrUpdateSLZ
from backend.templatesets.legacy_apps.instance.constants import K8S_CONFIGMAP_SYS_CONFIG
from backend.uniapps import utils as app_utils
from backend.uniapps.application.base_views import BaseAPI
from backend.uniapps.resource.constants import DEFAULT_SEARCH_FIELDS
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes
from backend.utils.renderers import BKAPIRenderer
from backend.utils.response import PermsResponse

from .base import ResourceOperate

logger = logging.getLogger(__name__)


class ConfigMaps(viewsets.ViewSet, BaseAPI, ResourceOperate):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    cate = 'K8sConfigMap'
    category = 'configmap'
    sys_config = K8S_CONFIGMAP_SYS_CONFIG

    def get_configmaps_by_cluster_id(self, request, params, project_id, cluster_id):
        """查询configmaps"""
        # 共享集群禁用该接口
        if get_cluster_type(cluster_id) == ClusterType.SHARED:
            return 0, []

        search_fields = copy.deepcopy(DEFAULT_SEARCH_FIELDS)

        search_fields.append("data.data")
        params.update({"field": ",".join(search_fields)})
        client = k8s.K8SClient(request.user.token.access_token, project_id, cluster_id, env=None)
        resp = client.get_configmap(params)

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
        """获取项目下所有的ConfigMap"""
        params = dict(request.GET.items())
        is_decode = request.GET.get('decode')
        is_decode = True if is_decode == '1' else False

        cluster_id = params['cluster_id']

        code, cluster_configmaps = self.get_configmaps_by_cluster_id(request, params, project_id, cluster_id)
        if code != ErrorCode.NoError:
            return Response({'code': code, 'message': cluster_configmaps})

        self.handle_data(
            cluster_configmaps,
            self.cate,
            cluster_id,
            is_decode,
            namespace_dict=app_utils.get_ns_id_map(request.user.token.access_token, project_id),
        )

        # 按时间倒序排列
        cluster_configmaps.sort(key=lambda x: x.get('createTime', ''), reverse=True)
        return PermsResponse(cluster_configmaps, NamespaceRequest(project_id=project_id, cluster_id=cluster_id))

    def delete_configmap(self, request, project_id, cluster_id, namespace, name):
        return self.delete_resource(request, project_id, cluster_id, namespace, name)

    def batch_delete_configmaps(self, request, project_id):
        return self.batch_delete_resource(request, project_id)

    def update_configmap(self, request, project_id, cluster_id, namespace, name):
        self.slz = K8sConfigMapCreateOrUpdateSLZ
        return self.update_resource(request, project_id, cluster_id, namespace, name)


class ConfigMapListView(viewsets.ViewSet):
    render_classes = (BKAPIRenderer,)

    def exist_list(self, request, project_id):
        """exist configmap list
        NOTE: perm is ignore in the stage
        """
        cluster_data = self._get_cluster_list(request, project_id)
        project_configmap_list = []
        for cluster in cluster_data:
            configmap_list = self._get_configmaps(request, project_id, cluster)
            project_configmap_list.extend(configmap_list)
        return Response({'data': project_configmap_list, 'code': ErrorCode.NoError})

    def _get_cluster_list(self, request, project_id):
        cluster_resp = paas_cc.get_all_clusters(request.user.token.access_token, project_id)
        if cluster_resp.get('code') != ErrorCode.NoError:
            raise error_codes.APIError.f(cluster_resp.get('message'))
        cluster_data = cluster_resp.get('data') or {}
        return cluster_data.get('results') or []

    def _get_configmaps(self, request, project_id, cluster):
        """get configmap from project and cluster"""
        fields = ','.join(['namespace', 'resourceName'])
        k8s_client = k8s.K8SClient(request.user.token.access_token, project_id, cluster['cluster_id'], env=None)
        configmap_resp = k8s_client.get_configmap(fields)
        if configmap_resp.get('code') != ErrorCode.NoError:
            logger.error('request bcs api error, %s' % configmap_resp.get('message'))
            return []
        data = configmap_resp.get('data') or []
        return [
            {
                'name': info['resourceName'],
                'namespace': info['namespace'],
                'cluster_id': cluster['cluster_id'],
                'cluster_name': cluster['name'],
            }
            for info in data
            if info['namespace'] not in K8S_SYS_NAMESPACE
        ]

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

from rest_framework import views, viewsets
from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.response import Response

from backend.components import paas_cc
from backend.container_service.projects.base.constants import LIMIT_FOR_ALL_DATA
from backend.iam.permissions.decorators import response_perms
from backend.iam.permissions.resources.namespace import NamespaceRequest, calc_iam_ns_id
from backend.iam.permissions.resources.namespace_scoped import (
    NamespaceScopedAction,
    NamespaceScopedPermCtx,
    NamespaceScopedPermission,
)
from backend.kube_core.hpa import constants, utils
from backend.resources.exceptions import DeleteResourceError
from backend.templatesets.legacy_apps.configuration.constants import K8sResourceName
from backend.uniapps.application.base_views import BaseAPI
from backend.uniapps.network.serializers import BatchResourceSLZ
from backend.uniapps.resource.views import ResourceOperate
from backend.utils.error_codes import error_codes
from backend.utils.renderers import BKAPIRenderer
from backend.utils.response import BKAPIResponse, PermsResponse

logger = logging.getLogger(__name__)


class HPA(viewsets.ViewSet, BaseAPI, ResourceOperate):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)
    category = K8sResourceName.K8sHPA.value
    iam_perm = NamespaceScopedPermission()

    @response_perms(
        action_ids=[NamespaceScopedAction.VIEW, NamespaceScopedAction.UPDATE, NamespaceScopedAction.DELETE],
        permission_cls=NamespaceScopedPermission,
        resource_id_key='iam_ns_id',
    )
    def list(self, request, project_id):
        """获取所有HPA数据"""
        cluster_id = request.query_params.get("cluster_id")
        hpa_list = utils.get_cluster_hpa_list(request, project_id, cluster_id)

        for h in hpa_list:
            h['iam_ns_id'] = calc_iam_ns_id(cluster_id, h['namespace'])

        return PermsResponse(hpa_list, NamespaceRequest(project_id=project_id, cluster_id=cluster_id))

    def check_namespace_use_perm(self, request, project_id, namespace_list):
        """检查是否有命名空间的使用权限"""
        access_token = request.user.token.access_token

        # 根据 namespace  查询 ns_id
        namespace_res = paas_cc.get_namespace_list(access_token, project_id, limit=LIMIT_FOR_ALL_DATA)
        namespace_data = namespace_res.get('data', {}).get('results') or []
        namespace_dict = {i['name']: i['id'] for i in namespace_data}
        namespace_cluster_dict = {i['name']: i['cluster_id'] for i in namespace_data}

        for namespace in namespace_list:
            # 检查是否有命名空间的使用权限
            perm_ctx = NamespaceScopedPermCtx(
                username=request.user.username,
                project_id=project_id,
                cluster_id=namespace_cluster_dict.get(namespace),
                name=namespace,
            )
            self.iam_perm.can_use(perm_ctx)
        return namespace_dict

    def delete(self, request, project_id, cluster_id, ns_name, name):
        username = request.user.username
        namespace_dict = self.check_namespace_use_perm(request, project_id, [ns_name])
        namespace_id = namespace_dict.get(ns_name)

        try:
            utils.delete_hpa(request, project_id, cluster_id, ns_name, namespace_id, name)
        except DeleteResourceError as error:
            message = "删除HPA:{}失败, [命名空间:{}], {}".format(name, ns_name, error)
            utils.activity_log(project_id, username, name, message, False)
            raise error_codes.APIError(message)

        message = "删除HPA:{}成功, [命名空间:{}]".format(name, ns_name)
        utils.activity_log(project_id, username, name, message, True)

        return Response({})

    def batch_delete(self, request, project_id):
        """批量删除资源"""
        username = request.user.username

        slz = BatchResourceSLZ(data=request.data)
        slz.is_valid(raise_exception=True)
        data = slz.data['data']
        # 检查用户是否有命名空间的使用权限
        namespace_list = set([_d.get('namespace') for _d in data])
        namespace_dict = self.check_namespace_use_perm(request, project_id, namespace_list)
        success_list = []
        failed_list = []
        for _d in data:
            cluster_id = _d.get('cluster_id')
            name = _d.get('name')
            ns_name = _d.get('namespace')
            ns_id = namespace_dict.get(ns_name)

            # 删除 hpa
            try:
                utils.delete_hpa(request, project_id, cluster_id, ns_name, ns_id, name)
            except DeleteResourceError as error:
                failed_list.append({'name': name, 'desc': "{}[命名空间:{}]:{}".format(name, ns_name, error)})
            else:
                success_list.append({'name': name, 'desc': "{}[命名空间:{}]".format(name, ns_name)})

        # 添加操作审计
        message = '--'
        if success_list:
            name_list = [_s.get('name') for _s in success_list]
            desc_list = [_s.get('desc') for _s in success_list]

            desc_list_msg = ";".join(desc_list)
            message = "以下HPA删除成功:{}".format(desc_list_msg)

            utils.activity_log(project_id, username, ';'.join(name_list), message, True)

        if failed_list:
            name_list = [_s.get('name') for _s in failed_list]
            desc_list = [_s.get('desc') for _s in failed_list]

            desc_list_msg = ";".join(desc_list)
            message = "以下HPA删除失败:{}".format(desc_list_msg)

            utils.activity_log(project_id, username, ';'.join(name_list), message, False)

        # 有一个失败，则返回失败
        if failed_list:
            raise error_codes.APIError(message)

        return BKAPIResponse({}, message=message)


class HPAMetrics(views.APIView):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get(self, request, project_id):
        """获取支持的HPA metric列表"""
        return Response(constants.HPA_METRICS)

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
from typing import List, Set

from rest_framework.response import Response

from backend.bcs_web.audit_log.audit.decorators import log_audit_on_view
from backend.bcs_web.audit_log.constants import ActivityType, ResourceType
from backend.bcs_web.viewsets import UserViewSet
from backend.container_service.clusters.open_apis.auditors import OpenAPIAuditor
from backend.container_service.clusters.open_apis.serializers import CreateNamespaceSLZ, UpdateNamespaceSLZ
from backend.container_service.clusters.permissions import AccessClusterPermMixin
from backend.resources.namespace import Namespace
from backend.resources.namespace import utils as ns_utils
from backend.resources.namespace.constants import BCS_RESERVED_NAMESPACES, K8S_PLAT_NAMESPACE
from backend.templatesets.var_mgmt.models import NameSpaceVariable
from backend.utils.error_codes import error_codes


class NamespaceViewSet(AccessClusterPermMixin, UserViewSet):

    lookup_field = 'ns_name'

    def list(self, request, project_id_or_code, cluster_id):
        namespaces = ns_utils.get_namespaces_by_cluster_id(
            request.user.token.access_token, request.project.project_id, cluster_id
        )
        return Response(namespaces)

    @log_audit_on_view(OpenAPIAuditor, activity_type=ActivityType.Add)
    def create(self, request, project_id_or_code, cluster_id):
        project_id = request.project.project_id
        params = self.params_validate(CreateNamespaceSLZ, context={'project_id': project_id})
        ns_name, variables = params['name'], params['variables']
        if ns_name in BCS_RESERVED_NAMESPACES:
            raise error_codes.ValidateError('不允许创建 BCS 保留的命名空间')
        # 更新操作审计信息
        request.audit_ctx.update_fields(resource_type=ResourceType.Namespace, resource=ns_name)

        namespace = Namespace(request.ctx_cluster).get_or_create_cc_namespace(
            ns_name, request.user.username, params['labels'], params['annotations']
        )
        # 创建命名空间下的变量值
        ns_id = namespace['namespace_id']
        namespace['id'] = ns_id
        NameSpaceVariable.batch_save(ns_id, variables)
        namespace['variables'] = variables
        return Response(namespace)

    def retrieve(self, request, project_id_or_code, cluster_id, ns_name):
        namespace = Namespace(request.ctx_cluster).get(ns_name, is_format=False)
        if not namespace:
            raise error_codes.ResNotFoundError('集群 {} 中不存在命名空间 {}'.format(cluster_id, ns_name))
        return Response(namespace.data.to_dict())

    @log_audit_on_view(OpenAPIAuditor, activity_type=ActivityType.Modify)
    def update(self, request, project_id_or_code, cluster_id, ns_name):
        if ns_name in BCS_RESERVED_NAMESPACES:
            raise error_codes.ValidateError('不允许更新 BCS 保留的命名空间')
        # 更新操作审计信息
        request.audit_ctx.update_fields(resource_type=ResourceType.Namespace, resource=ns_name)

        params = self.params_validate(UpdateNamespaceSLZ)
        ns_client = Namespace(request.ctx_cluster)
        namespace = ns_client.get(ns_name, is_format=False)
        if not namespace:
            raise error_codes.ResNotFoundError('集群 {} 中不存在命名空间 {}'.format(cluster_id, ns_name))

        manifest = namespace.data.to_dict()
        for key in ['labels', 'annotations']:
            manifest['metadata'][key] = params[key]
        ns_client.replace(name=ns_name, body=manifest)
        return Response(manifest)

    @log_audit_on_view(OpenAPIAuditor, activity_type=ActivityType.Modify)
    def sync_namespaces(self, request, project_id_or_code, cluster_id):
        """同步集群命名空间到 BCSCC"""
        request.audit_ctx.update_fields(
            resource_type=ResourceType.Cluster, resource=cluster_id, description=f'同步集群 {cluster_id} 命名空间'
        )

        project_id = request.project.project_id
        # 统一: 通过cc获取的数据，添加cc限制，区别于直接通过线上直接获取的命名空间
        cc_namespaces = ns_utils.get_namespaces_by_cluster_id(request.user.token.access_token, project_id, cluster_id)
        # 转换格式，方便其他系统使用
        cc_namespace_name_id = {info["name"]: info["id"] for info in cc_namespaces}
        # 获取线上的命名空间
        access_token = request.user.token.access_token
        namespaces = ns_utils.get_k8s_namespaces(access_token, project_id, cluster_id)
        # NOTE: 忽略k8s系统和平台自身的namespace
        namespace_name_list = [
            info["resourceName"] for info in namespaces if info["resourceName"] not in K8S_PLAT_NAMESPACE
        ]
        if not (cc_namespaces and namespaces):
            return Response()
        # 根据namespace和realtime namespace进行删除或创建
        # 删除命名空间
        delete_ns_name_list = set(cc_namespace_name_id.keys()) - set(namespace_name_list)
        delete_ns_id_list = [cc_namespace_name_id[name] for name in delete_ns_name_list]
        self._delete_cc_ns(request, project_id, cluster_id, delete_ns_id_list)

        # 向V0权限中心注册命名空间数据
        add_ns_name_list = set(namespace_name_list) - set(cc_namespace_name_id.keys())
        self._add_cc_ns(request, project_id, cluster_id, add_ns_name_list)
        return Response()

    def _add_cc_ns(self, request, project_id: str, cluster_id: str, ns_name_list: Set[str]):
        access_token = request.user.token.access_token
        creator = request.user.token.access_token
        for ns_name in ns_name_list:
            ns_utils.create_cc_namespace(access_token, project_id, cluster_id, ns_name, creator)

    def _delete_cc_ns(self, request, project_id: str, cluster_id: str, ns_id_list: List[int]):
        """删除存储在CC中的namespace"""
        for ns_id in ns_id_list:
            ns_utils.delete_cc_namespace(request.user.token.access_token, project_id, cluster_id, ns_id)

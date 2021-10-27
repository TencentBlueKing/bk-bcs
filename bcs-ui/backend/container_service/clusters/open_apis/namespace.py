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
from typing import Dict

from rest_framework.response import Response

from backend.accounts import bcs_perm
from backend.bcs_web.viewsets import UserViewSet
from backend.container_service.clusters.base.models import CtxCluster
from backend.container_service.clusters.open_apis.serializers import CreateNamespaceParamsSLZ
from backend.resources.namespace import Namespace
from backend.resources.namespace import utils as ns_utils
from backend.resources.namespace.constants import K8S_PLAT_NAMESPACE
from backend.templatesets.var_mgmt.models import NameSpaceVariable


class NamespaceViewSet(UserViewSet):
    def list_by_cluster_id(self, request, project_id_or_code, cluster_id):
        namespaces = ns_utils.get_namespaces_by_cluster_id(
            request.user.token.access_token, request.project.project_id, cluster_id
        )
        return Response(namespaces)

    def create_namespace(self, request, project_id_or_code, cluster_id):
        project_id = request.project.project_id
        slz = CreateNamespaceParamsSLZ(data=request.data, context={"project_id": project_id})
        slz.is_valid(raise_exception=True)
        data = slz.data

        access_token = request.user.token.access_token
        username = request.user.username

        namespace = self._create_kubernetes_namespace(access_token, username, project_id, cluster_id, data["name"])
        # 创建命名空间下的变量值
        ns_id = namespace.get("namespace_id") or namespace.get("id")
        namespace["id"] = ns_id
        NameSpaceVariable.batch_save(ns_id, data["variables"])
        namespace["variables"] = data["variables"]

        # 命名空间权限Client
        ns_perm_client = bcs_perm.Namespace(request, project_id, bcs_perm.NO_RES, cluster_id)
        ns_perm_client.register(namespace["id"], f"{namespace['name']}({cluster_id})")

        return Response(namespace)

    def sync_namespaces(self, request, project_id_or_code, cluster_id):
        """同步集群下命名空间
        NOTE: 先仅处理k8s类型
        """
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
        self.delete_cc_ns(request, project_id, cluster_id, delete_ns_id_list)

        # 向V0权限中心注册命名空间数据
        add_ns_name_list = set(namespace_name_list) - set(cc_namespace_name_id.keys())
        self.add_cc_ns(request, project_id, cluster_id, add_ns_name_list)

        return Response()

    def add_cc_ns(self, request, project_id, cluster_id, ns_name_list):
        access_token = request.user.token.access_token
        creator = request.user.token.access_token
        perm = bcs_perm.Namespace(request, project_id, bcs_perm.NO_RES, cluster_id)
        for ns_name in ns_name_list:
            data = ns_utils.create_cc_namespace(access_token, project_id, cluster_id, ns_name, creator)
            perm.register(data["id"], f"{ns_name}({cluster_id})")

    def delete_cc_ns(self, request, project_id, cluster_id, ns_id_list):
        """删除存储在CC中的namespace"""
        for ns_id in ns_id_list:
            perm = bcs_perm.Namespace(request, project_id, ns_id)
            perm.delete()
            ns_utils.delete_cc_namespace(request.user.token.access_token, project_id, cluster_id, ns_id)

    def _create_kubernetes_namespace(
        self,
        access_token: str,
        username: str,
        project_id: str,
        cluster_id: str,
        ns_name: str,
    ) -> Dict:
        # TODO: 需要注意需要迁移到权限中心V3，通过注册到V0权限中心的命名空间ID，反查命名空间名称、集群ID及项目ID
        # 连接集群创建命名空间
        ctx_cluster = CtxCluster.create(token=access_token, id=cluster_id, project_id=project_id)
        return Namespace(ctx_cluster).get_or_create_cc_namespace(ns_name, username)

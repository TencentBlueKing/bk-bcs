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
from rest_framework.response import Response

from backend.bcs_web.viewsets import SystemViewSet
from backend.dashboard.permissions import AccessClusterPermMixin
from backend.resources.custom_object import CustomResourceDefinition, get_cobj_client_by_crd
from backend.utils.error_codes import error_codes

from .serializers import BatchDeleteCustomObjectsSLZ, PatchCustomObjectScaleSLZ, PatchCustomObjectSLZ
from .utils import to_table_format


class CRDViewSet(AccessClusterPermMixin, SystemViewSet):
    def list(self, request, project_id, cluster_id):
        crd_client = CustomResourceDefinition(request.ctx_cluster)
        return Response(crd_client.list())


class CustomObjectViewSet(AccessClusterPermMixin, SystemViewSet):
    def list_custom_objects(self, request, project_id, cluster_id, crd_name):
        crd_client = CustomResourceDefinition(request.ctx_cluster)
        crd = crd_client.get(name=crd_name, is_format=False)
        if not crd:
            raise error_codes.ResNotFoundError(_("集群({})中未注册自定义资源({})").format(cluster_id, crd_name))

        cobj_client = get_cobj_client_by_crd(request.ctx_cluster, crd_name)
        cobj_list = cobj_client.list(namespace=request.query_params.get("namespace"))
        return Response(to_table_format(crd.data.to_dict(), cobj_list, cluster_id=cluster_id))

    def get_custom_object(self, request, project_id, cluster_id, crd_name, name):
        cobj_client = get_cobj_client_by_crd(request.ctx_cluster, crd_name)
        cobj_dict = cobj_client.get(namespace=request.query_params.get("namespace"), name=name)
        return Response(cobj_dict)

    def patch_custom_object(self, request, project_id, cluster_id, crd_name, name):
        serializer = PatchCustomObjectSLZ(data=request.data)
        serializer.is_valid(raise_exception=True)

        validated_data = serializer.validated_data
        cobj_client = get_cobj_client_by_crd(request.ctx_cluster, crd_name)
        cobj_client.patch(
            name=name,
            namespace=validated_data.get("namespace"),
            body=validated_data["body"],
            content_type=validated_data["patch_type"],
        )

        return Response()

    def patch_custom_object_scale(self, request, project_id, cluster_id, crd_name, name):
        """自定义资源扩缩容"""
        req_data = request.data.copy()
        req_data["crd_name"] = crd_name
        serializer = PatchCustomObjectScaleSLZ(data=req_data)
        serializer.is_valid(raise_exception=True)

        validated_data = serializer.validated_data
        cobj_client = get_cobj_client_by_crd(request.ctx_cluster, crd_name)
        cobj_client.patch(
            name=name,
            namespace=validated_data.get("namespace"),
            body=validated_data["body"],
            content_type=validated_data["patch_type"],
        )

        return Response()

    def delete_custom_object(self, request, project_id, cluster_id, crd_name, name):
        cobj_client = get_cobj_client_by_crd(request.ctx_cluster, crd_name)
        cobj_client.delete_ignore_nonexistent(namespace=request.query_params.get("namespace"), name=name)
        return Response()

    def batch_delete_custom_objects(self, request, project_id, cluster_id, crd_name):
        serializer = BatchDeleteCustomObjectsSLZ(data=request.data)
        serializer.is_valid(raise_exception=True)

        validated_data = serializer.validated_data
        cobj_client = get_cobj_client_by_crd(request.ctx_cluster, crd_name)

        failed_list = []
        namespace = validated_data["namespace"]
        for name in validated_data["cobj_name_list"]:
            try:
                cobj_client.delete_ignore_nonexistent(namespace=namespace, name=name)
            except Exception:
                failed_list.append(name)

        if failed_list:
            raise error_codes.APIError(_("部分资源删除失败，失败列表: {}").format(",".join(failed_list)))

        return Response()

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
from kubernetes.dynamic.exceptions import DynamicApiError
from rest_framework.response import Response

from backend.bcs_web.audit_log.audit.decorators import log_audit_on_view
from backend.bcs_web.audit_log.constants import ActivityType
from backend.bcs_web.viewsets import SystemViewSet
from backend.dashboard.auditor import DashboardAuditor
from backend.dashboard.exceptions import CreateResourceError, DeleteResourceError, UpdateResourceError
from backend.dashboard.permissions import validate_cluster_perm
from backend.dashboard.serializers import CreateResourceSLZ, ListResourceSLZ, UpdateResourceSLZ
from backend.dashboard.utils.resp import ListApiRespBuilder, RetrieveApiRespBuilder
from backend.dashboard.utils.web import gen_base_web_annotations
from backend.utils.basic import getitems
from backend.utils.response import BKAPIResponse
from backend.utils.url_slug import KUBE_NAME_REGEX


class ListAndRetrieveMixin:
    """ 查询类接口通用逻辑 """

    def list(self, request, project_id, cluster_id, namespace=None):
        params = self.params_validate(ListResourceSLZ)
        client = self.resource_client(request.ctx_cluster)
        response_data = ListApiRespBuilder(client, namespace=namespace, **params).build()
        # 补充页面信息注解，包含权限信息
        web_annotations = gen_base_web_annotations(request, project_id, cluster_id)
        return BKAPIResponse(response_data, web_annotations=web_annotations)

    def retrieve(self, request, project_id, cluster_id, namespace, name):
        client = self.resource_client(request.ctx_cluster)
        response_data = RetrieveApiRespBuilder(client, namespace, name).build()
        # 补充页面信息注解，包含权限信息
        web_annotations = gen_base_web_annotations(request, project_id, cluster_id)
        return BKAPIResponse(response_data, web_annotations=web_annotations)


class DestroyMixin:
    """ 删除类接口通用逻辑 """

    @log_audit_on_view(DashboardAuditor, activity_type=ActivityType.Delete)
    def destroy(self, request, project_id, cluster_id, namespace, name):
        # 操作类接口统一检查集群操作权限
        validate_cluster_perm(request, project_id, cluster_id)
        client = self.resource_client(request.ctx_cluster)
        request.audit_ctx.update_fields(
            resource_type=self.resource_client.kind.lower(), resource=f'{namespace}/{name}'
        )
        try:
            response_data = client.delete(name=name, namespace=namespace).to_dict()
        except DynamicApiError as e:
            raise DeleteResourceError(_('删除资源失败: {}').format(e.summary()))
        return Response(response_data)


class CreateMixin:
    """ 创建类接口通用逻辑 """

    @log_audit_on_view(DashboardAuditor, activity_type=ActivityType.Add)
    def create(self, request, project_id, cluster_id, namespace=None):
        # 操作类接口统一检查集群操作权限
        validate_cluster_perm(request, project_id, cluster_id)
        params = self.params_validate(CreateResourceSLZ)
        client = self.resource_client(request.ctx_cluster)
        namespace = namespace or getitems(params, 'manifest.metadata.namespace')
        request.audit_ctx.update_fields(
            resource_type=self.resource_client.kind.lower(),
            resource=f"{namespace}/{getitems(params, 'manifest.metadata.name')}",
        )
        try:
            response_data = client.create(namespace=namespace, body=params['manifest'], is_format=False).data.to_dict()
        except DynamicApiError as e:
            raise CreateResourceError(_('创建资源失败: {}').format(e.summary()))
        except ValueError as e:
            raise CreateResourceError(_('创建资源失败: {}').format(str(e)))

        return Response(response_data)


class UpdateMixin:
    """ 更新类接口通用逻辑 """

    @log_audit_on_view(DashboardAuditor, activity_type=ActivityType.Modify)
    def update(self, request, project_id, cluster_id, namespace, name):
        # 操作类接口统一检查集群操作权限
        validate_cluster_perm(request, project_id, cluster_id)
        params = self.params_validate(UpdateResourceSLZ)
        client = self.resource_client(request.ctx_cluster)
        request.audit_ctx.update_fields(
            resource_type=self.resource_client.kind.lower(), resource=f'{namespace}/{name}'
        )
        manifest = params['manifest']
        # replace 模式下无需指定 resourceVersion
        manifest['metadata'].pop('resourceVersion', None)
        try:
            response_data = client.replace(
                body=manifest, namespace=namespace, name=name, is_format=False
            ).data.to_dict()
        except DynamicApiError as e:
            raise UpdateResourceError(_('更新资源失败: {}').format(e.summary()))
        except ValueError as e:
            raise UpdateResourceError(_('更新资源失败: {}').format(str(e)))

        return Response(response_data)


class DashboardViewSet(ListAndRetrieveMixin, DestroyMixin, CreateMixin, UpdateMixin, SystemViewSet):
    """
    资源视图通用 ViewSet，抽层一些通用方法
    """

    lookup_field = 'name'
    lookup_value_regex = KUBE_NAME_REGEX

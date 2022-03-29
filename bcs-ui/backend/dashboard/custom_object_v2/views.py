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
from backend.dashboard.auditors import DashboardAuditor
from backend.dashboard.constants import DashboardAction
from backend.dashboard.custom_object_v2 import serializers as slzs
from backend.dashboard.custom_object_v2.permissions import AccessCustomObjectsPermission
from backend.dashboard.custom_object_v2.utils import gen_cobj_web_annotations
from backend.dashboard.exceptions import CreateResourceError, DeleteResourceError, UpdateResourceError
from backend.dashboard.permissions import AccessClusterPermMixin
from backend.dashboard.utils.resp import ListApiRespBuilder, RetrieveApiRespBuilder
from backend.dashboard.utils.web import gen_base_web_annotations
from backend.dashboard.viewsets import PermValidateMixin
from backend.resources.constants import K8sResourceKind
from backend.resources.custom_object import CustomResourceDefinition, get_cobj_client_by_crd
from backend.resources.custom_object.formatter import CustomObjectCommonFormatter
from backend.utils.basic import getitems
from backend.utils.response import BKAPIResponse
from backend.utils.url_slug import KUBE_NAME_REGEX


class CRDViewSet(AccessClusterPermMixin, PermValidateMixin, SystemViewSet):
    """ 自定义资源定义 """

    lookup_field = 'crd_name'
    # 指定符合 CRD 名称规范的
    lookup_value_regex = KUBE_NAME_REGEX

    def list(self, request, project_id, cluster_id):
        """ 获取所有自定义资源列表 """
        self._validate_perm(request.user.username, project_id, cluster_id, None, DashboardAction.View)
        client = CustomResourceDefinition(request.ctx_cluster)
        response_data = ListApiRespBuilder(client).build()
        return Response(response_data)

    def retrieve(self, request, project_id, cluster_id, crd_name):
        """ 获取单个自定义资源详情 """
        self._validate_perm(request.user.username, project_id, cluster_id, None, DashboardAction.View)
        client = CustomResourceDefinition(request.ctx_cluster)
        response_data = RetrieveApiRespBuilder(client, namespace=None, name=crd_name).build()
        return Response(response_data)


class CustomObjectViewSet(PermValidateMixin, SystemViewSet):
    """ 自定义资源对象 """

    lookup_field = 'custom_obj_name'
    lookup_value_regex = KUBE_NAME_REGEX

    def get_permissions(self):
        """ 在共享集群中仅部分自定义资源可订阅 """
        return [*super().get_permissions(), AccessCustomObjectsPermission()]

    def list(self, request, project_id, cluster_id, crd_name):
        """ 获取某类自定义资源列表 """
        params = self.params_validate(
            slzs.ListCustomObjectSLZ, context={'crd_name': crd_name, 'ctx_cluster': request.ctx_cluster}
        )
        client = get_cobj_client_by_crd(request.ctx_cluster, crd_name)
        namespace = params.get('namespace')
        self._validate_perm(request.user.username, project_id, cluster_id, namespace, DashboardAction.View)
        response_data = ListApiRespBuilder(
            client, formatter=CustomObjectCommonFormatter(), namespace=namespace
        ).build()
        web_annotations = gen_cobj_web_annotations(request, project_id, cluster_id, namespace, crd_name)
        return BKAPIResponse(response_data, web_annotations=web_annotations)

    def retrieve(self, request, project_id, cluster_id, crd_name, custom_obj_name):
        """ 获取单个自定义对象 """
        params = self.params_validate(
            slzs.FetchCustomObjectSLZ, context={'crd_name': crd_name, 'ctx_cluster': request.ctx_cluster}
        )
        client = get_cobj_client_by_crd(request.ctx_cluster, crd_name)
        namespace = params.get('namespace')
        self._validate_perm(request.user.username, project_id, cluster_id, namespace, DashboardAction.View)
        response_data = RetrieveApiRespBuilder(
            client, namespace=namespace, name=custom_obj_name, formatter=CustomObjectCommonFormatter()
        ).build()
        web_annotations = gen_base_web_annotations(request.user.username, project_id, cluster_id, namespace)
        return BKAPIResponse(response_data, web_annotations=web_annotations)

    @log_audit_on_view(DashboardAuditor, activity_type=ActivityType.Add)
    def create(self, request, project_id, cluster_id, crd_name):
        """ 创建自定义资源 """
        params = self.params_validate(slzs.CreateCustomObjectSLZ)
        namespace = getitems(params, 'manifest.metadata.namespace')
        cus_obj_name = getitems(params, 'manifest.metadata.name')
        self._validate_perm(request.user.username, project_id, cluster_id, namespace, DashboardAction.Create)
        self._update_audit_ctx(request, namespace, crd_name, cus_obj_name)

        client = get_cobj_client_by_crd(request.ctx_cluster, crd_name)
        try:
            response_data = client.create(namespace=namespace, body=params['manifest'], is_format=False).data.to_dict()
        except DynamicApiError as e:
            raise CreateResourceError(_('创建资源失败: {}').format(e.summary()))
        except ValueError as e:
            raise CreateResourceError(_('创建资源失败: {}').format(str(e)))

        return Response(response_data)

    @log_audit_on_view(DashboardAuditor, activity_type=ActivityType.Modify)
    def update(self, request, project_id, cluster_id, crd_name, custom_obj_name):
        """ 更新自定义资源 """
        params = self.params_validate(
            slzs.UpdateCustomObjectSLZ, context={'crd_name': crd_name, 'ctx_cluster': request.ctx_cluster}
        )
        namespace = params.get('namespace')
        self._validate_perm(request.user.username, project_id, cluster_id, namespace, DashboardAction.Update)
        self._update_audit_ctx(request, namespace, crd_name, custom_obj_name)

        client = get_cobj_client_by_crd(request.ctx_cluster, crd_name)
        manifest = params['manifest']
        # 自定义资源 Replace 也需要指定 resourceVersion
        # 这里先 pop，通过在 replace 指定 auto_add_version=True 添加
        manifest['metadata'].pop('resourceVersion', None)
        try:
            response_data = client.replace(
                body=manifest, namespace=namespace, name=custom_obj_name, is_format=False, auto_add_version=True
            ).data.to_dict()
        except DynamicApiError as e:
            raise UpdateResourceError(_('更新资源失败: {}').format(e.summary()))
        except ValueError as e:
            raise UpdateResourceError(_('更新资源失败: {}').format(str(e)))

        return Response(response_data)

    @log_audit_on_view(DashboardAuditor, activity_type=ActivityType.Delete)
    def destroy(self, request, project_id, cluster_id, crd_name, custom_obj_name):
        """ 删除自定义资源 """
        params = self.params_validate(
            slzs.DestroyCustomObjectSLZ, context={'crd_name': crd_name, 'ctx_cluster': request.ctx_cluster}
        )
        namespace = params.get('namespace')
        self._validate_perm(request.user.username, project_id, cluster_id, namespace, DashboardAction.Delete)
        self._update_audit_ctx(request, namespace, crd_name, custom_obj_name)

        client = get_cobj_client_by_crd(request.ctx_cluster, crd_name)
        try:
            response_data = client.delete(name=custom_obj_name, namespace=namespace).to_dict()
        except DynamicApiError as e:
            raise DeleteResourceError(_('删除资源失败: {}').format(e.summary()))
        return Response(response_data)

    @staticmethod
    def _update_audit_ctx(request, namespace: str, crd_name: str, custom_obj_name: str) -> None:
        """ 更新操作审计相关信息 """
        resource_name = (
            f'{crd_name} - {namespace}/{custom_obj_name}' if namespace else f'{crd_name} - {custom_obj_name}'
        )
        request.audit_ctx.update_fields(
            resource_type=K8sResourceKind.CustomObject.value.lower(), resource=resource_name
        )

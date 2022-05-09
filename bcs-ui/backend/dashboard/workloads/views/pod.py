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
from rest_framework.decorators import action
from rest_framework.response import Response

from backend.bcs_web.audit_log.audit.decorators import log_audit_on_view
from backend.bcs_web.audit_log.constants import ActivityType
from backend.dashboard.auditors import DashboardAuditor
from backend.dashboard.constants import DashboardAction
from backend.dashboard.exceptions import DeleteResourceError, OwnerReferencesNotExist
from backend.dashboard.viewsets import NamespaceScopeViewSet
from backend.resources.configs.configmap import ConfigMap
from backend.resources.configs.secret import Secret
from backend.resources.storages.persistent_volume_claim import PersistentVolumeClaim
from backend.resources.workloads.pod import Pod
from backend.utils.basic import getitems


class PodViewSet(NamespaceScopeViewSet):

    resource_client = Pod

    @action(methods=['GET'], url_path='pvcs', detail=True)
    def persistent_volume_claims(self, request, project_id, cluster_id, namespace, name):
        """获取 Pod Persistent Volume Claim 信息"""
        # 检查是否有查看命名空间域资源权限
        self._validate_perm(request.user.username, project_id, cluster_id, namespace, DashboardAction.View)
        response_data = Pod(request.ctx_cluster).filter_related_resources(
            PersistentVolumeClaim(request.ctx_cluster), namespace, name
        )
        return Response(response_data)

    @action(methods=['GET'], url_path='configmaps', detail=True)
    def configmaps(self, request, project_id, cluster_id, namespace, name):
        """获取 Pod ConfigMap 信息"""
        # 检查是否有查看命名空间域资源权限
        self._validate_perm(request.user.username, project_id, cluster_id, namespace, DashboardAction.View)
        response_data = Pod(request.ctx_cluster).filter_related_resources(
            ConfigMap(request.ctx_cluster), namespace, name
        )
        return Response(response_data)

    @action(methods=['GET'], url_path='secrets', detail=True)
    def secrets(self, request, project_id, cluster_id, namespace, name):
        """获取 Pod Secret 信息"""
        # 检查是否有查看命名空间域资源权限
        self._validate_perm(request.user.username, project_id, cluster_id, namespace, DashboardAction.View)
        response_data = Pod(request.ctx_cluster).filter_related_resources(Secret(request.ctx_cluster), namespace, name)
        return Response(response_data)

    @action(methods=['PUT'], url_path='reschedule', detail=True)
    @log_audit_on_view(DashboardAuditor, activity_type=ActivityType.Reschedule)
    def reschedule(self, request, project_id, cluster_id, namespace, name):
        """重新调度 Pod（仅对有父级资源的 Pod 有效）"""
        # 检查是否有更新命名空间域资源权限（重新调度视为更新操作）
        self._validate_perm(request.user.username, project_id, cluster_id, namespace, DashboardAction.Update)
        client = Pod(request.ctx_cluster)
        request.audit_ctx.update_fields(
            resource_type=self.resource_client.kind.lower(), resource=f'{namespace}/{name}'
        )

        # 检查 Pod 配置，必须有父级资源才可以重新调度
        pod_manifest = client.fetch_manifest(namespace, name)
        if not getitems(pod_manifest, 'metadata.ownerReferences'):
            raise OwnerReferencesNotExist(_('Pod {}/{} 不存在父级资源，无法被重新调度').format(namespace, name))
        # 重新调度的原理是直接删除 Pod，利用父级资源重新拉起服务
        try:
            response_data = client.delete(name=name, namespace=namespace).to_dict()
        except DynamicApiError as e:
            raise DeleteResourceError(_('重新调度 Pod 失败: {}').format(e.summary()))
        return Response(response_data)

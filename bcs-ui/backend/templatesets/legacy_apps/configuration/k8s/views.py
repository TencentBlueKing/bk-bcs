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

Done:
1. k8s 模本实例化后台转换的项目：
- env 环境变量转换 (环境变量前端统一存放在 webCache.env_list 中，有后台组装为 env & envFrom)
- resources 资源限制后添加单位
- 去掉前端字段 webCache
- 启动命令和参数用 shellhex 命令处理为数组
- lifecycle.command 用 shellhex 命令处理为数组
- hostNetwork 0/1 转换为 false/true
- 健康&就绪检查 type 存放在 curContainer.webCache.livenessProbeType/curContainer.webCache.readinessProbeType
- 健康&就绪检查 exec.command 多个参数用空格分隔，组装为数组后存储
- STATEFULSET->pvc.storage 添加单位Gi
4. json schema 验证 k8s 资源（constants_k8s.py）

TODO:
2. k8s 非标准日志采集
3. K8sDeployment/DaemonSet/Job/StatefulSet  Config 中端口名称/挂载名不能重复（serializers.py）
"""
import json

from django.utils.translation import ugettext_lazy as _
from rest_framework import viewsets
from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.response import Response

from backend.uniapps.application.constants import K8S_KIND
from backend.utils.error_codes import error_codes
from backend.utils.renderers import BKAPIRenderer

from ..constants import K8sResourceName
from ..mixins import GetVersionedEntity
from ..models import K8sService, K8sStatefulSet, get_k8s_container_ports, get_pod_qsets_by_tag


class TemplateResourceView(viewsets.ViewSet, GetVersionedEntity):
    """页面上依赖关系的API
    - K8sConfigMap 列表 （Deployment/DaemonSet/Job/StatefulSet 页面使用）
    - K8sSecret 列表 （Deployment/DaemonSet/Job/StatefulSet 页面使用）
    """

    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def list_configmaps(self, request, project_id, version_id):
        """查看模板集指定版本的configmap信息"""
        ventity = self.get_versioned_entity(project_id, version_id)
        return Response(ventity.get_configmaps_by_kind(K8S_KIND))

    def list_secrets(self, request, project_id, version_id):
        """查看模板集指定版本的secret信息"""
        ventity = self.get_versioned_entity(project_id, version_id)
        return Response(ventity.get_secrets_by_kind(K8S_KIND))

    def list_svc_selector_labels(self, request, project_id, version_id):
        """查看模板集指定版本的 label 信息"""
        ventity = self.get_versioned_entity(project_id, version_id)
        return Response(ventity.get_k8s_svc_selector_labels())

    def list_pod_resources(self, request, project_id, version_id):
        ventity = self.get_versioned_entity(project_id, version_id)
        return Response(ventity.get_k8s_pod_resources())

    def list_deployments(self, request, project_id, version_id):
        """查看模板集指定版本的k8s_deployment信息"""
        ventity = self.get_versioned_entity(project_id, version_id)
        return Response(ventity.get_k8s_deploys())

    def list_services(self, request, project_id, version_id):
        """查看模板集指定版本的k8s_service信息"""
        ventity = self.get_versioned_entity(project_id, version_id)
        return Response(ventity.get_k8s_services())

    def _get_tag_list(self, request):
        deploy_tag_list = request.GET.get('deploy_tag_list')
        try:
            tag_list = json.loads(deploy_tag_list)
        except Exception:
            raise error_codes.ValidateError(_("请选择关联的应用"))
        return tag_list

    def list_pod_res_labels(self, request, project_id, version_id):
        """查看模板集指定版本的label信息"""
        ventity = self.get_versioned_entity(project_id, version_id)
        tag_list = self._get_tag_list(request)

        pod_res_qsets = get_pod_qsets_by_tag(tag_list, ventity)
        if not pod_res_qsets:
            return Response({})

        label_map = pod_res_qsets[0].get_labels()

        for pod_res in pod_res_qsets[1:]:
            label = pod_res.get_labels()
            label_map = dict(label_map.items() & label.items())

        return Response(label_map)

    def list_container_ports(self, request, project_id, version_id):
        """查看模板集指定版本的端口信息"""
        ventity = self.get_versioned_entity(project_id, version_id)
        tag_list = self._get_tag_list(request)

        pod_res_qsets = get_pod_qsets_by_tag(tag_list, ventity)
        ports = get_k8s_container_ports(pod_res_qsets)
        return Response(ports)

    def check_port_associated_with_service(self, request, project_id, version_id, port_id):
        """检查指定的 port 是否被 service 关联"""
        ventity = self.get_versioned_entity(project_id, version_id)
        svc_id_list = ventity.get_resource_id_list(K8sResourceName.K8sService.value)
        svc_qsets = K8sService.objects.filter(id__in=svc_id_list)
        for svc in svc_qsets:
            ports = svc.get_ports_config()
            for p in ports:
                if str(p.get('id')) == str(port_id):
                    raise error_codes.APIError(_("端口在 Service[{}] 中已经被关联,不能删除").format(svc.name))
        return Response({})

    def update_sts_service_tag(self, request, project_id, version_id, sts_deploy_tag):
        service_tag = request.data.get('service_tag')
        ventity = self.get_versioned_entity(project_id, version_id)
        sts_id_list = ventity.get_resource_id_list(K8sResourceName.K8sStatefulSet.value)
        K8sStatefulSet.objects.filter(id__in=sts_id_list, deploy_tag=sts_deploy_tag).update(service_tag=service_tag)
        return Response({'sts_deploy_tag': sts_deploy_tag})

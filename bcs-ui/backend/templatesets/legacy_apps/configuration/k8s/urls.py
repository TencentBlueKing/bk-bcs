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
from django.conf.urls import url

from . import views

# ###### K8s 页面依赖的API
urlpatterns = [
    # 查询指定版本的  configmap 信息
    url(
        r'^api/configuration/(?P<project_id>\w{32})/K8sConfigMap/(?P<version_id>\-?\d+)/$',
        views.TemplateResourceView.as_view({'get': 'list_configmaps'}),
    ),
    # 查询指定版本的  secret 信息
    url(
        r'^api/configuration/(?P<project_id>\w{32})/K8sSecret/(?P<version_id>\-?\d+)/$',
        views.TemplateResourceView.as_view({'get': 'list_secrets'}),
    ),
    # 查询指定版本的 Deployment 信息
    url(
        r'^api/configuration/(?P<project_id>\w{32})/K8sDeployment/(?P<version_id>\-?\d+)/$',
        views.TemplateResourceView.as_view({'get': 'list_deployments'}),
    ),
    # 查询指定版本的 Pod 信息
    url(
        r'^api/configuration/(?P<project_id>\w{32})/pods/(?P<version_id>\-?\d+)/$',
        views.TemplateResourceView.as_view({'get': 'list_pod_resources'}),
    ),
    # 查询指定版本的已经被Service关联的label信息
    url(
        r'^api/configuration/(?P<project_id>\w{32})/labels/(?P<version_id>\-?\d+)/$',
        views.TemplateResourceView.as_view({'get': 'list_svc_selector_labels'}),
    ),
    # 查询指定版本的 K8sService 信息
    url(
        r'^api/configuration/(?P<project_id>\w{32})/K8sService/(?P<version_id>\-?\d+)/$',
        views.TemplateResourceView.as_view({'get': 'list_services'}),
    ),
    # 查询模板集指定版本中 Deployment/Statefulset等资源的container的port信息
    url(
        r'^api/configuration/projects/(?P<project_id>\w{32})/versions/(?P<version_id>\-?\d+)/K8sContainerPorts/$',
        views.TemplateResourceView.as_view({'get': 'list_container_ports'}),
    ),
    # 检查指定的 port 是否被 service 关联 TODO mark refactor 前端似乎未用
    url(
        r'^api/configuration/(?P<project_id>\w{32})/K8sDeployment/check/version/(?P<version_id>\d+)/port/'
        r'(?P<port_id>\-?\d+)/$',
        # noqa
        views.TemplateResourceView.as_view({'get': 'check_port_associated_with_service'}),
    ),
    # 查询指定deployment 中 label 信息
    url(
        r'^api/configuration/(?P<project_id>\w{32})/K8sDeployment/labels/(?P<version_id>\-?\d+)/$',
        views.TemplateResourceView.as_view({'get': 'list_pod_res_labels'}),
    ),
    url(
        r'^api/configuration/projects/(?P<project_id>\w{32})/versions/(?P<version_id>\-?\d+)/'
        r'K8sStatefulSet/(?P<sts_deploy_tag>\d{16})/service-tag/$',
        views.TemplateResourceView.as_view({'put': 'update_sts_service_tag'}),
    ),
]

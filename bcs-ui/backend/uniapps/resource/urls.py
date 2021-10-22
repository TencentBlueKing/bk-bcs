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

K8S_CLUSTER_ID_REGEX = "BCS-K8S-[0-9]{5,7}"


urlpatterns = [
    # ConfigMap：列表
    url(r'^api/resource/(?P<project_id>\w{32})/configmaps/$', views.ConfigMaps.as_view({'get': 'get'})),
    # Secrets：列表
    url(r'^api/resource/(?P<project_id>\w{32})/secrets/$', views.Secrets.as_view({'get': 'get'})),
    # ConfigMap 更新时展示的数据
    url(r'^api/resource/(?P<project_id>\w{32})/configmaps/update/$', views.ConfigMaps.as_view({'get': 'get'})),
    # Secrets 更新时展示的数据
    url(r'^api/resource/(?P<project_id>\w{32})/secrets/update/$', views.Secrets.as_view({'get': 'get'})),
    # endpoints：detail
    url(
        r'^api/resource/(?P<project_id>\w{32})/clusters/(?P<cluster_id>[\w\-]+)/namespaces/(?P<namespace>[\w\-]+)'
        '/endpoints/(?P<name>[\w.\-]+)/$',
        views.Endpoints.as_view(),
    ),
    # configmaps 操作
    url(
        r'^api/resource/(?P<project_id>\w{32})/configmaps/clusters/(?P<cluster_id>[\w\-]+)/'
        'namespaces/(?P<namespace>[\w\-]+)/endpoints/(?P<name>[\w.\-]+)/$',
        views.ConfigMaps.as_view({'post': 'update_configmap', 'delete': 'delete_configmap'}),
    ),
    # secrets 操作
    url(
        r'^api/resource/(?P<project_id>\w{32})/secrets/clusters/(?P<cluster_id>[\w\-]+)/'
        'namespaces/(?P<namespace>[\w\-]+)/endpoints/(?P<name>[\w.\-]+)/$',
        views.Secrets.as_view({'post': 'update_secret', 'delete': 'delete_secret'}),
    ),
    # 批量删除configmap
    url(
        r'^api/resource/(?P<project_id>\w{32})/configmaps/batch/$',
        views.ConfigMaps.as_view({'post': 'batch_delete_configmaps'}),
    ),
    # 批量删除 secrets
    url(
        r'^api/resource/(?P<project_id>\w{32})/secrets/batch/$',
        views.Secrets.as_view({'post': 'batch_delete_secrets'}),
    ),
    # Ingress 列表
    url(r'^api/resource/(?P<project_id>\w{32})/ingresses/$', views.ingress.IngressResource.as_view({'get': 'get'})),
    # 删除单个Ingress
    url(
        r'^api/resource/(?P<project_id>\w{32})/ingresses/clusters/(?P<cluster_id>[\w\-]+)/'
        'namespaces/(?P<namespace>[\w\-]+)/endpoints/(?P<name>[\w.\-]+)/$',
        views.ingress.IngressResource.as_view({'delete': 'delete_ingress'}),
    ),
    # 批量删除
    url(
        r'^api/resource/(?P<project_id>\w{32})/ingresses/batch/$',
        views.ingress.IngressResource.as_view({'post': 'batch_delete_ingress'}),
    ),
    # search exist configmap
    url(
        r'^api/resource/projects/(?P<project_id>\w{32})/configmap/exist/list/$',
        views.ConfigMapListView.as_view({'get': 'exist_list'}),
    ),
    url(
        r'^api/projects/(?P<project_id>\w{32})/clusters/(?P<cluster_id>%s)/'
        r'namespaces/(?P<namespace>[\w\-]+)/ingresses/(?P<name>[\w.\-]+)/$' % K8S_CLUSTER_ID_REGEX,
        views.ingress.IngressResource.as_view({"put": "update_ingress"}),
    ),
]

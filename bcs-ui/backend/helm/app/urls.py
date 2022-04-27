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
import os

from django.conf import settings
from django.conf.urls import url
from django.views.static import serve

from . import views

urlpatterns = [
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/apps/$',
        views.AppView.as_view({'get': 'list', 'post': 'create'}),
    ),
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/apps/create_preview/$',
        views.AppCreatePreviewView.as_view({'post': 'create'}),
    ),
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/apps/(?P<app_id>\d+)/release_preview/$',
        views.AppReleasePreviewView.as_view({'post': 'create'}),
    ),
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/apps/(?P<app_id>\d+)/rollback_preview/$',
        views.AppRollbackPreviewView.as_view({'post': 'create'}),
    ),
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/apps/(?P<app_id>\d+)/preview/$',
        views.AppPreviewView.as_view({'get': 'retrieve'}),
    ),
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/apps/(?P<app_id>\d+)/$',
        views.AppView.as_view({'get': 'retrieve', 'put': 'update', 'delete': 'destroy'}),
    ),
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/apps/(?P<app_id>\d+)/transitioning/$',
        views.AppTransiningView.as_view({'get': 'retrieve'}),
    ),
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/apps/(?P<app_id>\w+)/rollback/$',
        views.AppRollbackView.as_view({'get': 'retrieve', 'put': 'update'}),
    ),
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/namespaces/$',
        views.AppNamespaceView.as_view({'get': 'list'}),
    ),
    # 升级 app 时, 可选的版本, 包含一个特性选项，用于保持 app 的模板不变
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/apps/(?P<app_id>\d+)/upgrade_versions/$',
        views.AppUpgradeVersionsView.as_view({'get': 'list'}),
    ),
    # 回滚 app 时，可选的 release, 从应用的release列表中剔除了当前的release
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/apps/(?P<app_id>\d+)/rollback_selections/$',
        views.AppRollbackSelectionsView.as_view({'get': 'list'}),
    ),
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/apps/(?P<app_id>\d+)/differences/$',
        views.AppReleaseDiffView.as_view({'post': 'create'}),
    ),
    # 应用升级时：用于获取选择的升级目标版本内容，注意：这里面的 id 可能是 -1
    url(
        (
            r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/apps/(?P<app_id>\d+)/'
            'update_chart_versions/(?P<update_chart_version_id>(\-?\d+))/$'
        ),
        views.AppUpdateChartVersionView.as_view({'get': 'retrieve'}),
    ),
    # backend function as tools
    url(r'^api/bcs/k8s/tools/sync_dict2yaml/$', views.SyncDict2YamlToolView.as_view({'post': 'create'})),
    url(r'^api/bcs/k8s/tools/sync_yaml2dict/$', views.SyncYaml2DictToolView.as_view({'post': 'create'})),
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/apps/(?P<app_id>\d+)/state/$',
        views.AppStateView.as_view({'get': 'retrieve'}),
    ),
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/apps/(?P<app_id>\d+)/status/$',
        views.AppStatusView.as_view({'get': 'retrieve'}),
    ),
    url(
        r'^api/bcs/k8s/apps/(?P<app_id>\d+)/$',
        views.AppAPIView.as_view({'put': 'update'}),
        name="api.helm.app.update",
    ),
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/how-to-push-helm-chart/$',
        views.HowToPushHelmChartView.as_view({'get': 'retrieve'}),
    ),
    url(
        r'^api/bcs/k8s/documents/(?P<path>.*)/$',
        serve,
        {'document_root': os.path.join(settings.BASE_DIR, 'backend/helm/app/documentation/'), "show_indexes": True},
    ),
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/container/registry/domian/$',
        views.ContainerRegistryDomainView.as_view({'get': 'retrieve'}),
    ),
]

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

PROJECT_ID = "(?P<project_id>[\w\-]+)"
REPO_NAME = "(?P<repo_name>[a-z0-9_-]{1,32})"
REPO_ID = "(?P<repo_id>[0-9]+)"

urlpatterns = [
    # repository
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/helm/repositories/lists/detailed',
        views.RepositoryView.as_view({'get': 'list_detailed'}),
        name='api.helm.helm_repositories_list_detailed',
    ),
    # 用户可能并不关心 chart 属于那个 repo，只是想从所有的chart中找某个chart
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/helm/charts/$',
        views.ChartViewSet.as_view({"get": "list"}),
        name='api.helm.helm_repo_chart_list',
    ),
    # chart version
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/helm/repositories/(?P<repo_id>[0-9]+)/'
        'charts/(?P<chart_id>[0-9]+)/versions/$',
        views.ChartVersionView.as_view({'get': 'list'}),
        name='api.helm.helm_repo_chart_version_list',
    ),
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/helm/repositories/(?P<repo_id>[0-9]+)/'
        'charts/(?P<chart_id>[0-9]+)/versions/(?P<version_id>[0-9]+)/$',
        views.ChartVersionView.as_view({'get': 'retrieve'}),
        name='api.helm.helm_repo_chart_version_detail',
    ),
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/helm/' 'charts/(?P<chart_id>[0-9]+)/versions/$',
        views.ChartVersionView.as_view({'get': 'list'}),
        name='api.helm.helm_repo_chart_version_list',
    ),
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/helm/'
        'charts/(?P<chart_id>[0-9]+)/versions/(?P<version_id>[0-9]+)/$',
        views.ChartVersionView.as_view({'get': 'retrieve'}),
        name='api.helm.helm_repo_chart_version_detail',
    ),
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/helm/repositories/(?P<repo_id>[0-9]+)/sync/$',
        views.RepositorySyncView.as_view({'post': 'create'}),
        name='api.helm.helm_repositories_sync',
    ),
    url(
        r'^api/bcs/k8s/configuration/(?P<project_id>\w{32})/helm/repositories/sync/$',
        views.RepositorySyncByProjectView.as_view({'post': 'create'}),
        name='api.helm.helm_repositories_sync_by_project',
    ),
    url(
        r'^api/bcs/k8s/configuration_noauth/(?P<sync_project_id>\w{32})/helm/repositories/sync/$',
        views.RepositorySyncByProjectAPIView.as_view({'post': 'create'}),
        name='api.helm.helm_repositories_sync_by_project',
    ),
    url(
        r'^api/projects/(?P<project_id>\w{32})/helm/charts/(?P<chart_id>\d+)/releases/$',
        views.ChartVersionViewSet.as_view({"get": "release_list"}),
    ),
    url(
        r'^api/projects/(?P<project_id>\w{32})/helm/charts/(?P<chart_id>\d+)/$',
        views.ChartVersionViewSet.as_view({"delete": "delete"}),
    ),
    url(
        r'^api/projects/(?P<project_id>\w{32})/helm/charts/(?P<chart_name>[\w\-]+)/releases/$',
        views.HelmChartVersionsViewSet.as_view({"post": "list_releases_by_chart_versions"}),
    ),
    url(
        r'^api/projects/(?P<project_id>\w{32})/helm/charts/(?P<chart_name>[\w\-]+)/$',
        views.HelmChartVersionsViewSet.as_view({"delete": "batch_delete"}),
    ),
]

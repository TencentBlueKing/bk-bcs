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

urlpatterns = [
    url(r'^charts/$', views.ChartsApiView.as_view({'get': 'list_charts'})),
    url(
        r'^charts/(?P<chart_id>\d+)/$',
        views.ChartsApiView.as_view({'get': 'retrieve_chart'}),
    ),
    url(
        r'^charts/(?P<chart_id>\d+)/versions/$',
        views.ChartVersionApiView.as_view({'get': 'chart_versions'}),
    ),
    url(
        r'^charts/(?P<chart_id>\d+)/versions/(?P<version_id>\d+)/$',
        views.ChartVersionApiView.as_view({'get': 'retrieve_chart_version'}),
    ),
    url(
        r'^charts/(?P<chart_id>\d+)/versions/(?P<version_id>\d+)/valuefile/$',
        views.ChartVersionApiView.as_view({'get': 'retrieve_valuefile'}),
    ),
    url(
        r'^namespaces/$',
        views.ChartAppNamespaceApiView.as_view({'get': 'list_available_namespaces'}),
    ),
    url(
        r'^apps/$',
        views.ChartsAppApiView.as_view({'post': 'create_app', 'get': 'list_app'}),
    ),
    url(
        r'^apps/(?P<app_id>\d+)/transition/$',
        views.ChartsAppTransitionApiView.as_view({'get': 'retrieve_app'}),
    ),
    url(
        r'^apps/(?P<app_id>\d+)/$',
        views.ChartsAppApiView.as_view({'put': 'update', 'delete': 'delete_app', 'get': 'retrieve'}),
    ),
    url(
        r'^apps/(?P<app_id>\d+)/upgrade_versions/$',
        views.AppUpgradeVersionView.as_view({'get': 'list_app_versions'}),
    ),
    url(
        r'^repositories/sync/$',
        views.SyncRepoView.as_view({'post': 'sync_repo'}),
    ),
    url(
        r"^charts/(?P<chart_name>[\w\-]+)/$",
        views.DeleteChartOrVersion.as_view({"delete": "delete"}),
    ),
    url(r"^repo/$", views.ChartRepoViewSet.as_view({"get": "retrieve"})),
]

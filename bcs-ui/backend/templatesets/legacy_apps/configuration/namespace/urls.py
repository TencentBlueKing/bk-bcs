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
    # namespace 操作
    url(
        r'^api/configuration/(?P<project_id>\w{32})/namespace/$',
        views.NamespaceView.as_view({'get': 'list', 'post': 'create'}),
    ),
    url(
        r'^api/configuration/(?P<project_id>\w{32})/namespace/(?P<namespace_id>\d+)/$',
        views.NamespaceView.as_view({'put': 'update', 'delete': 'delete'}),
    ),
    url(
        r'^api/projects/(?P<project_id>\w{32})/configuration/namespaces/sync/$',
        views.NamespaceView.as_view({'post': 'sync_namespace'}),
    ),
    url(
        r'^api/projects/(?P<project_id>\w{32})/namespaces/(?P<namespace_id>\d+)/resources/$',
        views.NamespaceView.as_view({"get": "get_ns_resources"}),
    ),
    url(
        r"^api/resources/projects/(?P<project_id>\w{32})/"
        "clusters/(?P<cluster_id>[\w-]+)/namespaces/(?P<namespace>[\w-]+)/$",
        views.NamespaceQuotaViewSet.as_view(
            {"get": "get_namespace_quota", "put": "update_namespace_quota", "delete": "delete_namespace_quota"}
        ),
    ),
]

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
from django.conf.urls import include, url

from backend.helm.open_apis.views import SharedChartRepoViewSet
from backend.web_console.open_apis.views import WebConsoleSession

urlpatterns = [
    url(r"^projects/", include("backend.container_service.projects.open_apis.urls")),
    # TODO ^resources/projects/ will replace ^projects/(?P<project_id>\w{32})/clusters/ in apigw
    url(
        r"^projects/(?P<project_id_or_code>[\w\-]+)/clusters/",
        include("backend.container_service.clusters.open_apis.urls"),
    ),
    url(
        r"^resources/projects/(?P<project_id_or_code>[\w\-]+)/clusters/",
        include("backend.container_service.clusters.open_apis.urls"),
    ),
    # TODO ^templatesets/projects/ will replace ^projects/(?P<project_id>\w{32})/configuration/ in apigw
    url(r"^projects/(?P<project_id_or_code>[\w\-]+)/configuration/", include("backend.templatesets.open_apis.urls")),
    url(r"^templatesets/projects/(?P<project_id_or_code>[\w\-]+)/", include("backend.templatesets.open_apis.urls")),
    # TODO ^templatesets/projects/ will replace ^projects/(?P<project_id>\w{32})/configuration/templates/ in apigw
    url(
        r"^projects/(?P<project_id>\w{32})/configuration/templates/",
        include("backend.templatesets.open_apis.template_urls"),
    ),
    url(
        r"^templatesets/projects/(?P<project_id>\w{32})/templates/",
        include("backend.templatesets.open_apis.template_urls"),
    ),
    # 提供给iam拉取资源实例的url(已注册到iam后台)
    url(r"^iam/", include("backend.iam.open_apis.urls")),
    # web_console API
    url(
        r"^projects/(?P<project_id_or_code>[\w\-]+)/clusters/(?P<cluster_id>[\w\-]+)/web_console/sessions/",
        WebConsoleSession.as_view(),
    ),
    # TODO ^helm/projects/ will replace ^projects/(?P<project_id_or_code>[\w\-]+)/helm/ in apigw
    url(r"^projects/(?P<project_id_or_code>[\w\-]+)/helm/", include("backend.helm.open_apis.urls")),
    url(r"^helm/projects/(?P<project_id_or_code>[\w\-]+)/", include("backend.helm.open_apis.urls")),
    url(r"^helm/public_repo/$", SharedChartRepoViewSet.as_view({"get": "retrieve"})),
    url(r"^var_mgmt/projects/(?P<project_id>\w{32})/", include("backend.templatesets.var_mgmt.open_apis.urls")),
]

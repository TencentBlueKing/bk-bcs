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
    # project list
    url(r"^api/projects/$", views.Projects.as_view({"get": "list"})),
    # compatible with project_id, project_code
    url(
        r"^api/projects/(?P<project_id>[\w-]+)/$",
        views.Projects.as_view({"get": "info", "put": "update"}),
        name="update_project",
    ),
    # get cmdb business
    url(r"^api/cc/$", views.CC.as_view({"get": "list"})),
    url(r"^api/nav/users/$", views.UserAPIView.as_view()),
    url(
        r"^api/nav/projects/$", views.NavProjectsViewSet.as_view({"get": "filter_projects", "post": "create_project"})
    ),
    url(
        r"^api/nav/projects/(?P<project_id>\w{32})/$",
        views.NavProjectsViewSet.as_view({"get": "get_project", "put": "update_project"}),
    ),
    url(r"^api/nav/projects/user_perms/$", views.NavProjectPermissionViewSet.as_view({"post": "get_user_perms"})),
    url(
        r"^api/nav/projects/(?P<project_id>\w{32})/user_perms/$",
        views.NavProjectPermissionViewSet.as_view({"post": "query_user_perms_by_project"}),
    ),
    url(
        r"^api/projects/(?P<project_id>\w{32})/biz_maintainers/$",
        views.ProjectBizInfoViewSet.as_view({"get": "list_biz_maintainers"}),
    ),
]

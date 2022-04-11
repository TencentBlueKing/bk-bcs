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
    url(r"^api/authorized_projects/$", views.AuthorizedProjectsViewSet.as_view({'get': 'list'})),
    # compatible with project_id, project_code
    url(
        r"^api/projects/(?P<project_id>[\w-]+)/$",
        views.Projects.as_view({"get": "get_project", "put": "update_bound_biz"}),
        name="update_project",
    ),
    url(
        r"^api/projects/(?P<project_id>\w{32})/biz_maintainers/$",
        views.ProjectBizInfoViewSet.as_view({"get": "list_biz_maintainers"}),
    ),
    # get cmdb business
    url(r"^api/cc/$", views.CC.as_view({"get": "list"})),
    # nav 用于私有化版本的项目管理功能
    url(r"^api/nav/users/$", views.UserAPIView.as_view()),
    url(r"^api/nav/projects/$", views.NavProjectsViewSet.as_view({'get': 'list_projects', "post": "create_project"})),
    url(
        r"^api/nav/projects/(?P<project_id>\w{32})/$",
        views.NavProjectsViewSet.as_view({"get": "get_project", "put": "update_project"}),
    ),
]

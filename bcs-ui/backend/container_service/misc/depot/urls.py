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
    # 镜像库
    url(r'^api/depot/images/public/$', views.SharedRepoImagesViewSet.as_view({"get": "get"})),
    # 项目镜像
    url(
        r'^api/depot/images/project/(?P<project_id>\w{32})/$',
        views.ProjectImagesViewSet.as_view({"get": "get"}),
        name='project_images',
    ),
    # 镜像库 + 项目镜像 : 提供给模板配置页面使用
    url(
        r'^api/depot/available/images/(?P<project_id>\w{32})/$',
        views.AvailableImagesViewSet.as_view({"get": "get"}),
        name='available_image',
    ),
    # 根据 镜像标识（repo） 查询 tags
    url(
        r'^api/depot/available/tags/(?P<project_id>\w{32})/$',
        views.AvailableTagsViewSet.as_view({"get": "get"}),
        name='available_tag',
    ),
    # 镜像详情API
    url(
        r'^api/depot/images/project/(?P<project_id>\w{32})/info/image/$',
        views.ImageDetailViewSet.as_view({'get': 'get'}),
    ),
]

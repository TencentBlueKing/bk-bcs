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
    # 查询模板下面已经实例化过的版本信息
    url(
        r'^api/configuration/(?P<project_id>\w{32})/template/exist_versions/(?P<template_id>\d+)/$',
        views.TemplateInstView.as_view({'get': 'get_exist_version'}),
    ),
    # 查询模板已经实例化过的版本的名称
    url(
        r'^api/configuration/(?P<project_id>\w{32})/exist/show_version_name/(?P<template_id>\d+)/$',
        views.TemplateInstView.as_view({'get': 'get_exist_showver_name'}),
    ),
    # 查询模板已经实例化过的版本下的资源
    url(
        r'^api/configuration/(?P<project_id>\w{32})/exist/resource/(?P<template_id>\d+)/$',
        views.TemplateInstView.as_view({'get': 'get_resource_by_show_name'}),
    ),
    # 模板实例化配置：创建 API
    url(r'^api/configuration/(?P<project_id>\w{32})/instances/$', views.VersionInstanceView.as_view({'post': 'post'})),
    # 预览配置文件
    url(
        r'^api/configuration/(?P<project_id>\w{32})/preview/$',
        views.VersionInstanceView.as_view({'post': 'preview_config'}),
    ),
    # 看下模板版本下已经实例化过的命名空间
    url(
        r'^api/configuration/(?P<project_id>\w{32})/instance/ns/(?P<version_id>\d+)/$',
        views.InstanceNameSpaceView.as_view({'post': 'post'}),
    ),
]

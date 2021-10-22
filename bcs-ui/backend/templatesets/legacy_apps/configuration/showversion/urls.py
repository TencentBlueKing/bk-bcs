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
    # 保存用户可见版本
    url(
        r'^api/configuration/(?P<project_id>\w{32})/template/(?P<template_id>\d+)/show/version/$',
        views.ShowVersionViewSet.as_view({'get': 'list_show_versions', 'post': 'save_with_ventity'}),
    ),
    # 根据模板集查询用户可见版本列表/ 创建可见版本
    url(
        r'^api/configuration/(?P<project_id>\w{32})/show/versions/(?P<template_id>\d+)/$',
        views.ShowVersionViewSet.as_view({'get': 'list_show_versions_for_instance', 'post': 'save_without_ventity'}),
    ),
    # 加载版本的内容
    url(
        r'^api/configuration/(?P<project_id>\w{32})/template/(?P<template_id>\d+)/show/version/'
        r'(?P<show_version_id>\-?\d+)/$',
        views.ShowVersionViewSet.as_view({'get': 'get_resource_config', 'delete': 'delete_show_version'}),
    ),
    # TODO replace "api/configuration/(?P<project_id>\w{32})/show/versions/(?P<template_id>\d+)/$"
    url(
        r'^api/projects/(?P<project_id>\w{32})/configuration/templates/(?P<template_id>\d+)/show_versions/$',
        views.ShowVersionViewSet.as_view({'get': 'list_show_versions'}),
    ),
]

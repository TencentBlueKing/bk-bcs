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
    url(
        r'^api/projects/(?P<project_id>\w{32})/configuration/yaml_templates/$',
        views.YamlTemplateViewSet.as_view({'post': 'create_template'}),
    ),
    url(
        r'^api/projects/(?P<project_id>\w{32})/configuration/yaml_templates/(?P<template_id>\d+)/$',
        views.YamlTemplateViewSet.as_view({'get': 'get_template', 'put': 'update_template'}),
    ),
    url(
        r'^api/projects/(?P<project_id>\w{32})/configuration/yaml_templates/(?P<template_id>\d+)/'
        r'show_versions/(?P<show_version_id>\d+)/$',
        views.YamlTemplateViewSet.as_view({'get': 'get_template_by_show_version'}),
    ),
    url(
        r'^api/projects/(?P<project_id>\w{32})/configuration/yaml_templates/(?P<template_id>\d+)/'
        r'show_versions/(?P<show_version_id>\d+)/releases/$',
        views.TemplateReleaseViewSet.as_view({'post': 'preview_or_apply'}),
    ),
    url(
        r'^api/projects/(?P<project_id>\w{32})/configuration/yaml_templates/initial_templates/$',
        views.InitialTemplatesViewSet.as_view({'get': 'get_initial_templates'}),
    ),
]

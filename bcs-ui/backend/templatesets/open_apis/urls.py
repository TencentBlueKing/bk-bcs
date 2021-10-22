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

from .form_mode import views as form_views
from .yaml_mode import views as yaml_views

urlpatterns = [
    url(
        r"^form_templates/(?P<template_id>\d+)/show_versions/(?P<show_version_id>\d+)/$",
        form_views.TemplateViewSet.as_view({"get": "get_template_by_show_version"}),
    ),
    url(
        r"^form_templates/(?P<template_id>\d+)/show_versions/(?P<show_version_id>\d+)/releases/$",
        form_views.TemplateReleaseViewSet.as_view({"post": "create_release"}),
    ),
    url(
        r"^form_templates/(?P<template_id>\d+)/releases/(?P<release_id>\d+)/$",
        form_views.TemplateReleaseViewSet.as_view({"put": "update_resource"}),
    ),
    url(
        r"^form_templates/(?P<template_id>\d+)/releases/$",
        form_views.TemplateReleaseViewSet.as_view({"get": "list_releases"}),
    ),
    url(
        r"^form_templates/(?P<template_id>\d+)/releases/latest/$",
        form_views.TemplateReleaseViewSet.as_view({"get": "get_latest_release"}),
    ),
    url(r"^yaml_templates/releases/$", yaml_views.TemplateReleaseViewSet.as_view({"post": "apply"})),
]

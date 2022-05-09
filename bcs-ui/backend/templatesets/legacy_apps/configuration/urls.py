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

from ..configuration.instance.urls import urlpatterns as inst_patterns
from ..configuration.k8s.urls import urlpatterns as k8s_patterns
from ..configuration.namespace.urls import urlpatterns as ns_patterns
from ..configuration.showversion.urls import urlpatterns as sversion_patterns
from ..configuration.yaml_mode.urls import urlpatterns as yaml_patterns
from . import views

urlpatterns = [
    # 模板集：列表/创建 API
    url(r'^api/configuration/(?P<project_id>\w{32})/templates/$', views.TemplatesView.as_view()),
    # 模板：查询／修改 API
    url(
        r'^api/configuration/(?P<project_id>\w{32})/template/(?P<pk>\d+)/$',
        views.SingleTemplateView.as_view(),
    ),
    url(
        r'^api/configuration/(?P<project_id>\w{32})/template/lock/(?P<template_id>\d+)/$',
        views.LockTemplateView.as_view({'post': 'lock_template'}),
    ),
    url(
        r'^api/configuration/(?P<project_id>\w{32})/template/unlock/(?P<template_id>\d+)/$',
        views.LockTemplateView.as_view({'post': 'unlock_template'}),
    ),
    # 保存草稿
    url(
        r'^api/configuration/(?P<project_id>\w{32})/template/(?P<template_id>\d+)/draft/$',
        views.CreateTemplateDraftView.as_view({'post': 'create_draft'}),
    ),
    # 资源 : 创建
    url(
        r'^api/configuration/(?P<project_id>\w{32})/template/(?P<template_id>\d+)/(?P<resource_name>\w+)/$',
        views.CreateAppResourceView.as_view(),
    ),
    # 资源：更新/删除，即创建新的版本
    url(
        r'^api/configuration/(?P<project_id>\w{32})/version/(?P<version_id>\d+)/(?P<resource_name>\w+)/'
        r'(?P<resource_id>\d+)/$',
        views.UpdateDestroyAppResourceView.as_view(),
    ),
    # 根据版本id查询所有的资源id
    url(r'^api/configuration/(?P<project_id>\w{32})/resource/(?P<pk>\d+)/$', views.TemplateResourceView.as_view()),
]

urlpatterns += sversion_patterns

urlpatterns += ns_patterns

# ###### K8s 页面依赖的API
urlpatterns += k8s_patterns

urlpatterns += inst_patterns

urlpatterns += yaml_patterns

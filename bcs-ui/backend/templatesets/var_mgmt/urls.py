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
    # 变量：列表/创建 API
    url(r'^api/configuration/(?P<project_id>\w{32})/variables/$', views.ListCreateVariableView.as_view()),
    # 变量：批量删除 API
    url(
        r'^api/configuration/(?P<project_id>\w{32})/variables/batch/$',
        views.VariableOverView.as_view({"delete": "batch_delete", "post": "batch_import"}),
    ),
    # 变量：查询／修改 API
    url(
        r'^api/configuration/(?P<project_id>\w{32})/variable/(?P<pk>\d+)/$', views.RetrieveUpdateVariableView.as_view()
    ),
    # 查询集群变量信息
    url(
        r'^api/configuration/(?P<project_id>\w{32})/variable/cluster/(?P<cluster_id>[\w\-]+)/$',
        views.ClusterVariableView.as_view({"get": "get_variables", "post": "batch_save"}),
    ),
    # 查询命名空间变量信息
    url(
        r'^api/configuration/(?P<project_id>\w{32})/variable/namespace/(?P<ns_id>\d+)/$',
        views.NameSpaceVariableView.as_view({"get": "get_variables"}),
    ),
    # 查询模板版本所有的变量信息
    url(
        r'^api/configuration/(?P<project_id>\w{32})/variable/resource/(?P<version_id>\d+)/$',
        views.ResourceVariableView.as_view(),
    ),
    # 查询变量在所有命名空间上的值
    url(
        r'^api/configuration/(?P<project_id>\w{32})/variable/batch/namespace/(?P<var_id>\d+)/$',
        views.NameSpaceVariableView.as_view({"get": "get_batch_variables", 'post': 'save_batch_variables'}),
    ),
    # 查询变量在所有集群上的值
    url(
        r'^api/configuration/(?P<project_id>\w{32})/variable/batch/cluster/(?P<var_id>\d+)/$',
        views.ClusterVariableView.as_view({"get": "get_batch_variables", 'post': 'save_batch_variables'}),
    ),
]

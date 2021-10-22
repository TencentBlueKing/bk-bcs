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

from backend.uniapps.apis.applications import views

urlpatterns = [
    url(
        r"cc_app_ids/(?P<cc_app_id>\d+)/projects/(?P<project_id>\w{32})/instances/$",
        views.ProjectApplicationInfo.as_view({"get": "get", "post": "api_post"}),
    ),
    url(
        r"cc_app_ids/(?P<cc_app_id>\d+)/projects/(?P<project_id>\w{32})/instance/detail/$",
        views.InstanceInfo.as_view({"get": "get"}),
    ),
    url(
        r"cc_app_ids/(?P<cc_app_id>\d+)/projects/(?P<project_id>\w{32})/instances/(?P<instance_name>[\w\-\.]+)/namespaces/$",  # noqa
        views.InstanceNamespace.as_view({"get": "get"}),
    ),
    url(
        r"cc_app_ids/(?P<cc_app_id>\d+)/projects/(?P<project_id>\w{32})/instances/(?P<instance_id>\d+)/retry/$",
        views.CreateInstance.as_view({"put": "api_put"}),
    ),
    url(
        r"cc_app_ids/(?P<cc_app_id>\d+)/projects/(?P<project_id>\w{32})/instances/(?P<instance_id>\d+)/update/$",
        views.UpdateInstance.as_view({"post": "api_post"}),
    ),
    url(
        r"cc_app_ids/(?P<cc_app_id>\d+)/projects/(?P<project_id>\w{32})/instances/batch_update/$",
        views.BatchUpdateInstance.as_view({"put": "api_put"}),
    ),
    url(
        r"cc_app_ids/(?P<cc_app_id>\d+)/projects/(?P<project_id>\w{32})/instances/(?P<instance_id>\d+)/scale/$",
        views.ScaleInstance.as_view({"post": "api_post"}),
    ),
    url(
        r"cc_app_ids/(?P<cc_app_id>\d+)/projects/(?P<project_id>\w{32})/instances/batch_scale/$",
        views.BatchScaleInstance.as_view({"put": "api_put"}),
    ),
    url(
        r"cc_app_ids/(?P<cc_app_id>\d+)/projects/(?P<project_id>\w{32})/instances/(?P<instance_id>\d+)/recreate/$",
        views.RecreateInstance.as_view({"put": "api_put"}),
    ),
    url(
        r"cc_app_ids/(?P<cc_app_id>\d+)/projects/(?P<project_id>\w{32})/instances/batch_recreate/$",
        views.BatchRecreateInstance.as_view({"post": "api_post"}),
    ),
    url(
        r"cc_app_ids/(?P<cc_app_id>\d+)/projects/(?P<project_id>\w{32})/instances/(?P<instance_id>\d+)/cancel/$",
        views.CancelInstance.as_view({"post": "api_post"}),
    ),
    url(
        r"cc_app_ids/(?P<cc_app_id>\d+)/projects/(?P<project_id>\w{32})/instances/(?P<instance_id>\d+)/pause/$",
        views.PauseInstance.as_view({"post": "api_post"}),
    ),
    url(
        r"cc_app_ids/(?P<cc_app_id>\d+)/projects/(?P<project_id>\w{32})/instances/(?P<instance_id>\d+)/resume/$",
        views.ResumeInstance.as_view({"post": "api_post"}),
    ),
    url(
        r"cc_app_ids/(?P<cc_app_id>\d+)/projects/(?P<project_id>\w{32})/instances/(?P<instance_id>\d+)/status/$",
        views.InstanceStatus.as_view({"get": "get"}),
    ),
    url(
        r"cc_app_ids/(?P<cc_app_id>\d+)/projects/(?P<project_id>\w{32})/instances/(?P<instance_id>\d+)/versions/$",
        views.GetInstanceVersions.as_view({"get": "get"}),
    ),
    url(
        r"cc_app_ids/(?P<cc_app_id>\d+)/projects/(?P<project_id>\w{32})/instances/(?P<instance_id>\d+)/version/configs/$",  # noqa
        views.GetInstanceVersionConf.as_view({"get": "get"}),
    ),
    url(
        r"cc_app_ids/(?P<cc_app_id>\d+)/projects/(?P<project_id>\w{32})/instances/batch_delete/$",  # noqa
        views.BatchDeleteInstance.as_view({"delete": "api_delete"}),
    ),
    url(
        r"cc_app_ids/(?P<cc_app_id>\d+)/projects/(?P<project_id>\w{32})/musters/$",  # noqa
        views.ProjectMuster.as_view({"get": "get"}),
    ),
    url(
        r"cc_app_ids/(?P<cc_app_id>\d+)/projects/(?P<project_id>\w{32})/musters/(?P<muster_id>\d+)/versions/$",  # noqa
        views.ProjectMusterVersion.as_view({"get": "get"}),
    ),
    url(
        r"cc_app_ids/(?P<cc_app_id>\d+)/projects/(?P<project_id>\w{32})/versions/(?P<version_id>\d+)/templates/$",  # noqa
        views.ProjectMusterTemplate.as_view({"get": "get"}),
    ),
    url(
        r"cc_app_ids/(?P<cc_app_id>\d+)/projects/(?P<project_id>\w{32})/namespaces/$",  # noqa
        views.ProjectNamespace.as_view({"post": "post"}),
    ),
]

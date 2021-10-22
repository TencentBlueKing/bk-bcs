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
from django.conf.urls import include, url
from rest_framework import routers

from . import views

router = routers.DefaultRouter(trailing_slash=True)

router.register(r'pods/(?P<pod_name>[\w\-.]+)/containers', views.ContainerMetricViewSet, basename='container')
router.register(r'pods', views.PodMetricViewSet, basename='pod')
router.register(r'nodes', views.NodeMetricViewSet, basename='node')
router.register(r'targets', views.TargetsViewSet, basename='target')
router.register(
    r'service_monitors/(?P<namespace>[\w-]+)', views.ServiceMonitorDetailViewSet, basename='service_monitor_detail'
)
router.register(r'service_monitors', views.ServiceMonitorViewSet, basename='service_monitor')
router.register(r'services', views.ServiceViewSet, basename='service')
router.register(r'', views.ClusterMetricViewSet, basename='cluster')

urlpatterns = [
    url(r'', include(router.urls)),
    # TODO 后续接入统一监控，不再需要该接口
    url(r"^prometheus/update/$", views.prometheus.PrometheusUpdateViewSet.as_view({"get": "get", "put": "update"})),
]

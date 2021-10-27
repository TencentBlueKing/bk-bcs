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
from rest_framework import routers

from . import views

router = routers.DefaultRouter(trailing_slash=True)

router.register(r'cronjobs', views.CronJobViewSet, basename='cronjob')
router.register(r'daemonsets', views.DaemonSetViewSet, basename='daemonset')
router.register(r'deployments', views.DeploymentViewSet, basename='deployment')
router.register(r'jobs', views.JobViewSet, basename='job')
router.register(r'pods', views.PodViewSet, basename='pod')
router.register(r'pods/(?P<pod_name>[\w\-.]+)/containers', views.ContainerViewSet, basename='container')
router.register(r'statefulsets', views.StatefulSetViewSet, basename='statefulset')

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

from backend.utils.url_slug import KUBE_NAME_REGEX, NAMESPACE_REGEX

from .cluster import ClusterDiscovererCacheViewSet, ClusterViewSet
from .deployment import DeploymentViewSet
from .namespace import NamespaceViewSet
from .pod import PodViewSet

router = routers.DefaultRouter(trailing_slash=True)
router.register('', NamespaceViewSet, basename='open_apis.namespace')

urlpatterns = [
    url(r"^$", ClusterViewSet.as_view({"get": "list"})),
    url(
        r"^(?P<cluster_id>[\w\-]+)/discoverer_cache/$",
        ClusterDiscovererCacheViewSet.as_view({"delete": "invalidate"}),
    ),
    url(r"^(?P<cluster_id>[\w\-]+)/crds/", include("backend.container_service.clusters.open_apis.custom_object.urls")),
    url(r'^(?P<cluster_id>[\w\-]+)/namespaces/', include(router.urls)),
    url(
        r"^(?P<cluster_id>[\w\-]+)/sync_namespaces/$",
        NamespaceViewSet.as_view({"put": "sync_namespaces"}),
    ),
    url(
        r"^(?P<cluster_id>[\w\-]+)/namespaces/(?P<namespace>%s)/deployments/$" % NAMESPACE_REGEX,
        DeploymentViewSet.as_view({"get": "list_by_namespace"}),
    ),
    url(
        r"^(?P<cluster_id>[\w\-]+)/namespaces/(?P<namespace>%s)/deployments/(?P<deploy_name>%s)/pods/$"
        % (NAMESPACE_REGEX, KUBE_NAME_REGEX),
        DeploymentViewSet.as_view({"get": "list_pods_by_deployment"}),
    ),
    url(
        r"^(?P<cluster_id>[\w\-]+)/namespaces/(?P<namespace>%s)/pods/(?P<pod_name>%s)/$"
        % (NAMESPACE_REGEX, KUBE_NAME_REGEX),
        PodViewSet.as_view({"get": "get_pod"}),
    ),
]

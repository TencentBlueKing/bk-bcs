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
from django.urls import include

from . import views
from .cc_host.urls import cc_router
from .featureflag.views import ClusterFeatureFlagViewSet
from .views.clusters import ClusterViewSet
from .views.node_views import nodes

urlpatterns = [
    # TODO: 老版本的集群查询
    url(r"^api/projects/(?P<project_id>\w+)/clusters/?$", ClusterViewSet.as_view({"get": "list"})),
    # 监控信息
    url(
        r'^api/projects/(?P<project_id>\w+)/metrics/cluster/summary/$',
        views.ClusterSummaryMetrics.as_view({'get': 'list'}),
    ),
    url(
        r'^api/projects/(?P<project_id>\w+)/metrics/cluster/?$',
        views.ClusterMetrics.as_view({'get': 'list'}),
    ),
    url(
        r'^api/projects/(?P<project_id>\w+)/metrics/node/?$',
        views.NodeMetrics.as_view({'get': 'list'}),
    ),
    url(
        r'^api/projects/(?P<project_id>\w+)/metrics/docker/?$',
        views.DockerMetrics.as_view({'get': 'list', 'post': 'multi'}),
    ),
    url(
        r'^api/projects/(?P<project_id>\w+)/metrics/node/summary/?$',
        views.NodeSummaryMetrics.as_view({'get': 'list'}),
    ),
    url(
        r"^api/projects/(?P<project_id>\w{32})/clusters/(?P<cluster_id>[\w\-]+)/masters/$",
        nodes.MasterViewSet.as_view({"get": "list"}),
    ),
]

# 新版 CC Host 相关接口
urlpatterns += [
    url(r'^api/projects/(?P<project_id>[\w\-]+)/cc/', include(cc_router.urls)),
]

# batch operation
urlpatterns += [
    url(
        r'^api/projects/(?P<project_id>[\w\-]+)/clusters/(?P<cluster_id>[\w\-]+)/pods/reschedule/$',
        nodes.BatchReschedulePodsViewSet.as_view({"put": "reschedule"}),
    ),
]

urlpatterns += [
    url(
        r"^api/cluster_mgr/projects/(?P<project_id>\w{32})/clusters/(?P<cluster_id>[\w\-]+)/nodes/$",
        nodes.NodeViewSets.as_view({"get": "list_nodes"}),
    )
]

# 节点 taint 相关 API
urlpatterns += [
    url(
        r"^api/cluster_mgr/projects/(?P<project_id>\w{32})/clusters/(?P<cluster_id>[\w\-]+)/nodes/taints/$",
        nodes.NodeViewSets.as_view({"post": "query_taints", "put": "set_taints"}),
    )
]

# 节点 标签 相关 API
urlpatterns += [
    url(
        r"^api/cluster_mgr/projects/(?P<project_id>\w{32})/clusters/(?P<cluster_id>[\w\-]+)/nodes/labels/$",
        nodes.NodeViewSets.as_view({"post": "query_labels", "put": "set_labels"}),
    )
]

urlpatterns += [
    url(
        r'^api/projects/(?P<project_id>[\w\-]+)/clusters/(?P<cluster_id>[\w\-]+)/feature_flags/$',
        ClusterFeatureFlagViewSet.as_view({'get': 'get_cluster_feature_flags'}),
    )
]

# 节点调度相关
urlpatterns += [
    url(
        r"^api/cluster_mgr/projects/(?P<project_id>\w{32})/clusters/(?P<cluster_id>[\w\-]+)/nodes/schedule_status/$",
        nodes.NodeViewSets.as_view({"put": "set_schedule_status"}),
    )
]

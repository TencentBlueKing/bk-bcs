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

from backend.dashboard.configs.urls import router as config_router
from backend.dashboard.custom_object_v2.urls import router as custom_obj_router
from backend.dashboard.examples.urls import router as example_router
from backend.dashboard.hpa.urls import router as hpa_router
from backend.dashboard.namespaces.urls import router as namespace_router
from backend.dashboard.networks.urls import router as network_router
from backend.dashboard.rbac.urls import router as rbac_router
from backend.dashboard.storages.urls import router as storage_router
from backend.dashboard.subscribe.urls import router as subscribe_router
from backend.dashboard.workloads.urls import router as workload_router

# 可选 namespaces/:namespace 前缀的 urls 集合
namespace_prefix_urlpatterns = [
    url(r"^configs/", include(config_router.urls)),
    url(r"^hpa/", include(hpa_router.urls)),
    url(r"^networks/", include(network_router.urls)),
    url(r"^rbac/", include(rbac_router.urls)),
    url(r"^storages/", include(storage_router.urls)),
    url(r"^workloads/", include(workload_router.urls)),
]

urlpatterns = [
    # TODO 自定义资源暂时保持 V1，V2 两个版本，后续会逐步替换掉 V1
    url(r"^crds/v2/", include(custom_obj_router.urls)),
    url(r"^crds/", include("backend.dashboard.custom_object.urls")),
    url(r"^namespaces/", include(namespace_router.urls)),
    url(r"^subscribe/", include(subscribe_router.urls)),
    url(r"^examples/", include(example_router.urls)),
    url(r"^(namespaces/(?P<namespace>[\w\-.]+)/)?", include(namespace_prefix_urlpatterns)),
]

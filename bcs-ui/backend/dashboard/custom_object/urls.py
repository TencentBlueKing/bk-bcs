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

from .views import CRDViewSet, CustomObjectViewSet

urlpatterns = [
    url(r"^$", CRDViewSet.as_view({"get": "list"})),
    url(
        r"^(?P<crd_name>[\w\.]+)/custom_objects/$",
        CustomObjectViewSet.as_view({"get": "list_custom_objects", "delete": "batch_delete_custom_objects"}),
    ),
    url(
        r"^(?P<crd_name>[\w\.]+)/custom_objects/(?P<name>[\w\-]+)/$",
        CustomObjectViewSet.as_view(
            {"get": "get_custom_object", "patch": "patch_custom_object", "delete": "delete_custom_object"}
        ),
    ),
    url(
        r"^(?P<crd_name>[\w\.]+)/custom_objects/(?P<name>[\w\-]+)/scale/$",
        CustomObjectViewSet.as_view({"patch": "patch_custom_object_scale"}),
    ),
]

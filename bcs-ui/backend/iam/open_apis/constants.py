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
from backend.iam.permissions.resources import constants
from backend.packages.blue_krill.data_types.enum import StructuredEnum

ResourceType = constants.ResourceType


class MethodType(str, StructuredEnum):
    """
    权限中心拉取资源的 method 参数值
    字段协议说明 https://bk.tencent.com/docs/document/6.0/160/8427?r=1
    """

    LIST_ATTR = 'list_attr'
    LIST_ATTR_VALUE = 'list_attr_value'
    LIST_INSTANCE = 'list_instance'
    FETCH_INSTANCE_INFO = 'fetch_instance_info'
    LIST_INSTANCE_BY_POLICY = 'list_instance_by_policy'
    SEARCH_INSTANCE = 'search_instance'

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
from backend.packages.blue_krill.data_types.enum import EnumField, StructuredEnum


class ToolStatus(str, StructuredEnum):
    PENDING = EnumField('pending', label='pending')
    DEPLOYED = EnumField('deployed', label='deployed')
    FAILED = EnumField('failed', label='failed')
    UNKNOWN = EnumField('unknown', label='unknown')


class OpType(str, StructuredEnum):
    INSTALL = 'install'
    UPGRADE = 'upgrade'
    UNINSTALL = 'uninstall'

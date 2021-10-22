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
from rest_framework import serializers

from backend.uniapps.apis.base_serializers import BaseParamsSLZ


class BaseProjectAPPParamsSZL(BaseParamsSLZ):
    category = serializers.CharField(required=True)


class ProjectAPPParamsSLZ(BaseProjectAPPParamsSZL):
    namespace = serializers.CharField(required=False)


class ProjectTemplateSetParamsSLZ(BaseParamsSLZ):
    pass


class BaseProjectParamsSLZ(BaseParamsSLZ):
    pass


class InstanceNamespaceParamsSLZ(BaseProjectAPPParamsSZL):
    pass


class InstanceStatusParamsSLZ(BaseParamsSLZ):
    pass


class ScaleInstanceParamsSLZ(BaseParamsSLZ):
    pass


class CreateInstanceParamsSLZ(BaseParamsSLZ):
    pass


class UpdateInstanceParamsSLZ(BaseParamsSLZ):
    pass


class BatchReCreateInstanceParamsSLZ(BaseParamsSLZ):
    pass


class BatchScaleInstanceParamsSLZ(BaseParamsSLZ):
    pass


class BatchUpdateInstanceParamsSLZ(BaseParamsSLZ):
    pass


class BatchDeleteInstanceParamsSLZ(BaseParamsSLZ):
    pass

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
from django.utils.translation import ugettext_lazy as _
from rest_framework import serializers
from rest_framework.exceptions import ValidationError

from backend.resources.constants import PatchType

from .constants import SupportedScaleCRDs


class PatchCustomObjectSLZ(serializers.Serializer):
    namespace = serializers.CharField(required=False)
    body = serializers.JSONField()
    patch_type = serializers.ChoiceField(choices=PatchType.get_choices(), default=PatchType.MERGE_PATCH_JSON.value)


class PatchCustomObjectScaleSLZ(PatchCustomObjectSLZ):
    crd_name = serializers.ChoiceField(choices=SupportedScaleCRDs.get_choices())

    def validate_body(self, body):
        try:
            replicas = int(body["spec"]["replicas"])
            if replicas < 0:
                raise ValueError(".spec.replicas must be a non-negative integer")
        except KeyError:
            raise ValidationError(_('body 参数格式不合法, 合法格式如 {"spec": {"replicas": 1}}'))
        except TypeError:
            raise ValidationError(_('body 参数格式不合法, 合法格式如 {"spec": {"replicas": 1}}'))
        except ValueError:
            raise ValidationError(_(".spec.replicas 必须设置为一个非负整数"))
        return body


class BatchDeleteCustomObjectsSLZ(serializers.Serializer):
    namespace = serializers.CharField()
    cobj_name_list = serializers.ListField(child=serializers.CharField(), min_length=1)

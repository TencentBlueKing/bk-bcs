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
from rest_framework.exceptions import ValidationError

from backend.templatesets.var_mgmt.models import Variable
from backend.templatesets.var_mgmt.serializers import RE_KEY


class VariablesSLZ(serializers.Serializer):
    id = serializers.IntegerField(required=False)
    key = serializers.RegexField(RE_KEY, max_length=64)
    value = serializers.CharField(default="")


class CreateNamespaceParamsSLZ(serializers.Serializer):
    name = serializers.RegexField(r'[a-z0-9]([-a-z0-9]*[a-z0-9])?', min_length=2, max_length=63)
    variables = serializers.ListField(child=VariablesSLZ(), default=[])

    def validate_variables(self, variables):
        if not variables:
            return variables
        project_id = self.context["project_id"]
        # ns_vars 格式 [{"name": "", "key": "", "value": ""}]
        var_key_list = [var["key"] for var in variables]
        var_qs = Variable.objects.filter(project_id=project_id, key__in=var_key_list)
        key_id_map = {var.key: var.id for var in var_qs}
        # 判断key是否存在
        not_exist_keys = set(var_key_list) - set(key_id_map.keys())
        if not_exist_keys:
            raise ValidationError(f"KEY: {','.join(not_exist_keys)} not found")
        # 添加对应的
        for var in variables:
            var["id"] = key_id_map[var["key"]]

        return variables

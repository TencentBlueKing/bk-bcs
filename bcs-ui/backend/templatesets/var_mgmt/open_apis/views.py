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
from rest_framework.response import Response

from backend.bcs_web.viewsets import UserViewSet

from . import var_helper


class VariablesViewSet(UserViewSet):
    def list_namespaced_variables(self, request, project_id, cluster_id, namespace):
        """获取命名空间下的变量"""
        # 获取命名空间ID
        ns_id = var_helper.get_ns_id(request.user.token.access_token, project_id, cluster_id, namespace)
        # 获取变量ID和命名空间下值的映射
        var_id_data_map = var_helper.get_var_data(ns_id)
        # 获取变量ID和对应信息的映射
        var_id_key_name_map = var_helper.get_var_key_and_name(var_id_data_map.keys())
        # 匹配数据，返回命名空间下的变量列表及对应的值
        return Response(var_helper.compose_data(var_id_data_map, var_id_key_name_map))

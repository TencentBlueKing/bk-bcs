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
import json

from rest_framework.response import Response

from backend.templatesets.legacy_apps.configuration.models import Template
from backend.templatesets.legacy_apps.configuration.namespace.views import NamespaceView
from backend.templatesets.legacy_apps.instance.models import InstanceConfig, VersionInstance
from backend.uniapps.apis.applications.views import BaseHandleInstance
from backend.uniapps.apis.base_serializers import BaseParamsSLZ
from backend.uniapps.apis.utils import skip_authentication
from backend.utils.renderers import BKAPIRenderer


class NamespaceApiView(BaseHandleInstance, NamespaceView):
    renderer_classes = (BKAPIRenderer,)

    def get_used_namespace(self, project_id):
        # 获取模板集ID
        template_id_list = Template.objects.filter(project_id=project_id, is_deleted=False).values_list(
            "id", flat=True
        )

        # 获取实例版本信息
        version_instance_id_list = VersionInstance.objects.filter(
            template_id__in=template_id_list, is_deleted=False
        ).values_list("id", flat=True)
        # 获取实例命名空间
        instance_namespace_list = InstanceConfig.objects.filter(
            instance_id__in=version_instance_id_list, is_deleted=False
        ).values_list("namespace", flat=True)
        return instance_namespace_list

    def get_all_ns_api(self, request, cc_app_id, project_id):
        self.init_handler(request, cc_app_id, project_id, 0, BaseParamsSLZ)
        used_namespace_flag = request.GET.get("used")
        project_namespace = self.list(request, project_id)
        if used_namespace_flag:
            used_namespace_list = self.get_used_namespace(project_id)
            content = json.loads(project_namespace.content)
            data = content.get("data") or []
            ret_data = []
            for info in data:
                if str(info["id"]) in used_namespace_list:
                    ret_data.append(info)
            return Response(ret_data)
        else:
            return project_namespace

    def get_ns_api(self, request, cc_app_id, project_id, namespace_id):
        self.init_handler(request, cc_app_id, project_id, 0, BaseParamsSLZ)
        return self.get_ns(request, project_id, namespace_id)

    def create_ns_api(self, request, cc_app_id, project_id):
        self.init_handler(request, cc_app_id, project_id, 0, BaseParamsSLZ)
        app_code = request.user.username
        is_validate_perm = True
        if skip_authentication(app_code):
            is_validate_perm = False
        return self.create(request, project_id, is_validate_perm)

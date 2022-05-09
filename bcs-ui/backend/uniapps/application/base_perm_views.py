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
from backend.accounts import bcs_perm
from backend.components import paas_cc
from backend.iam.permissions.resources.namespace_scoped import NamespaceScopedPermCtx, NamespaceScopedPermission
from backend.iam.permissions.resources.templateset import TemplatesetPermCtx, TemplatesetPermission
from backend.templatesets.legacy_apps.configuration.models import Template
from backend.utils.error_codes import error_codes

from .constants import SourceType


class InstancePerm:
    @classmethod
    def can_use_instance(cls, request, project_id, ns_id, tmpl_set_id=None, source_type=SourceType.TEMPLATE):
        # 继承命名空间的权限
        resp = paas_cc.get_namespace(request.user.token.access_token, project_id, ns_id)
        data = resp.get('data')
        perm_ctx = NamespaceScopedPermCtx(
            username=request.user.username,
            project_id=project_id,
            cluster_id=data.get('cluster_id'),
            name=data.get('name'),
        )
        NamespaceScopedPermission().can_use(perm_ctx)

        if source_type == SourceType.TEMPLATE:
            tmpl_set_info = Template.objects.filter(id=tmpl_set_id).first()
            if not tmpl_set_info:
                raise error_codes.CheckFailed(f"template:{tmpl_set_id} not found")
            # 继承模板集的权限
            perm_ctx = TemplatesetPermCtx(
                username=request.user.username, project_id=project_id, template_id=tmpl_set_id
            )
            TemplatesetPermission().can_instantiate(perm_ctx)

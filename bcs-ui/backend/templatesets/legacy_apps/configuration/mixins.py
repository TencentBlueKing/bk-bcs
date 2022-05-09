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
from backend.iam.permissions.resources.templateset import TemplatesetPermCtx, TemplatesetPermission

from .serializers_new import VentityWithTemplateSLZ
from .utils import validate_template_locked


class TemplatePermission:
    iam_perm = TemplatesetPermission()

    def can_edit_template(self, request, template):
        # 验证模板是否被其他用户加锁
        validate_template_locked(template, request.user.username)

        perm_ctx = TemplatesetPermCtx(
            username=request.user.username, project_id=template.project_id, template_id=template.id
        )
        self.iam_perm.can_update(perm_ctx)

    def can_view_template(self, request, template):
        # 验证用户是否有查看权限
        perm_ctx = TemplatesetPermCtx(
            username=request.user.username, project_id=template.project_id, template_id=template.id
        )
        self.iam_perm.can_view(perm_ctx)

    def can_use_template(self, request, template):
        # 验证用户是否有使用权限
        perm_ctx = TemplatesetPermCtx(
            username=request.user.username, project_id=template.project_id, template_id=template.id
        )
        self.iam_perm.can_instantiate(perm_ctx)

    def validate_template_locked(self, request, template):
        validate_template_locked(template, request.user.username)


class GetVersionedEntity:
    def get_versioned_entity(self, project_id, version_id):
        data = {'project_id': project_id, 'version_id': version_id}
        serializer = VentityWithTemplateSLZ(data=data)
        serializer.is_valid(raise_exception=True)
        return serializer.validated_data['ventity']

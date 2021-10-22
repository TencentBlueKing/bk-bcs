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
from backend.templatesets.legacy_apps.configuration.constants import TemplateEditMode
from backend.templatesets.models import AppReleaseData

from .form_mode import FormtoResourceList
from .res_context import ResContext
from .yaml_mode import YamltoResourceList

ResourceListGenerator = {
    TemplateEditMode.PageForm.value: FormtoResourceList,
    TemplateEditMode.YAML.value: YamltoResourceList,
}


class ReleaseDataGenerator:
    def __init__(self, name: str, res_ctx: ResContext):
        self.name = name
        self.res_ctx = res_ctx
        self.template = res_ctx.template
        self.generator = ResourceListGenerator[self.template.edit_mode]

    def generate(self) -> AppReleaseData:
        return AppReleaseData(
            name=self.name,
            project_id=self.res_ctx.project_id,
            cluster_id=self.res_ctx.cluster_id,
            namespace=self.res_ctx.namespace,
            template_id=self.template.id,
            resource_list=self.generator(self.res_ctx).generate(),
        )

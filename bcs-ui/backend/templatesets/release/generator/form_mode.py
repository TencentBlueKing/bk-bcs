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
from typing import List

from backend.templatesets.legacy_apps.instance.generator import GENERATOR_DICT
from backend.templatesets.legacy_apps.instance.utils import get_ns_variable
from backend.templatesets.models import ResourceData
from backend.utils.basic import getitems

from .res_context import ResContext


class FormtoResourceList:
    """表单模板集资源转换生成List[ResourceData]"""

    def __init__(self, res_ctx: ResContext):
        self.res_ctx = res_ctx

    # TODO 重构apps/instance模块, 挪到当前templatesets模块下
    def generate(self) -> List[ResourceData]:
        """
        先复用apps/instance/utils.generate_namespace_config中的大部分逻辑，完成resource_list的生成
        """
        res_ctx = self.res_ctx

        namespace_id = res_ctx.namespace_id
        # 查询命名空间相关的参数, 系统变量等保存在context中
        has_image_secret, cluster_version, var_context = get_ns_variable(
            res_ctx.access_token, res_ctx.project_id, namespace_id
        )
        show_version = res_ctx.show_version
        # 先按原apps/instance/generator.py中的方式组织params
        params = {
            'access_token': res_ctx.access_token,
            'project_id': res_ctx.project_id,
            'username': res_ctx.username,
            'instance_id': "0",
            'show_version_id': show_version.id,
            'version': show_version.name,
            'version_id': show_version.real_version_id,
            'template_id': show_version.template_id,
            'has_image_secret': has_image_secret,
            'cluster_version': cluster_version,
            'context': var_context,
            'variable_dict': res_ctx.template_variables,
            'is_preview': res_ctx.is_preview,
        }

        return self._generate(namespace_id, params)

    def _generate(self, namespace_id, params) -> List[ResourceData]:
        resource_list = []
        # instance_entity like {"Deployment": [1, 2]}
        instance_entity = self.res_ctx.instance_entity
        for kind in instance_entity:
            for entity_id in instance_entity[kind]:
                config_generator = GENERATOR_DICT.get(kind)(entity_id, namespace_id, is_validate=True, **params)
                config = config_generator.get_config_profile()
                try:
                    manifest = json.loads(config)
                except Exception:
                    manifest = config

                resource_list.append(
                    ResourceData(
                        kind=manifest.get('kind'),
                        name=getitems(manifest, 'metadata.name'),
                        namespace=getitems(manifest, 'metadata.namespace'),
                        manifest=manifest,
                        version=self.res_ctx.show_version.name,
                        revision=self.res_ctx.show_version.latest_revision,
                    )
                )

        return resource_list

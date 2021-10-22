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
import datetime
from typing import Dict, List

import jinja2
from rest_framework.exceptions import ParseError

from backend.helm.app import bcs_info_injector
from backend.helm.helm import bcs_variable
from backend.templatesets.legacy_apps.configuration.constants import FileResourceName
from backend.templatesets.legacy_apps.configuration.yaml_mode import res2files
from backend.templatesets.models import ResourceData
from backend.utils.basic import getitems

from .res_context import ResContext

ClusterScopedResources = [
    FileResourceName.ClusterRole.value,
    FileResourceName.ClusterRoleBinding.value,
    FileResourceName.StorageClass.value,
    FileResourceName.PersistentVolume.value,
]


class YamltoResourceList:
    """YAML模板集资源转换生成List[ResourceData]"""

    def __init__(self, res_ctx: ResContext):
        self.res_ctx = res_ctx

    def generate(self) -> List[ResourceData]:
        inject_configs = self._get_inject_configs()
        bcs_variables = self._get_bcs_variables()

        # template_variables是运行时变量，类似于Helm的--set
        template_variables = self.res_ctx.template_variables
        if template_variables:
            bcs_variables.update(template_variables)

        return self._generate(inject_configs, bcs_variables)

    def _generate(self, inject_configs: List[Dict], bcs_variables: Dict[str, str]) -> List[ResourceData]:
        """注入变量、渲染系统配置到原始的manifest中，最终生成待下发的ResourceData列表"""
        resource_list = []

        for raw_manifest in self._get_raw_manifests():
            rendered_manifest = self._render_with_variables(raw_manifest, bcs_variables)

            for manifest in self._inject_bcs_info(rendered_manifest, inject_configs):
                self._set_namespace(manifest)

                resource_list.append(
                    ResourceData(
                        kind=manifest.get('kind'),
                        name=getitems(manifest, 'metadata.name'),
                        namespace=self.res_ctx.namespace,
                        manifest=manifest,
                        version=self.res_ctx.show_version.name,
                        revision=self.res_ctx.show_version.latest_revision,
                    )
                )

        return resource_list

    def _get_inject_configs(self) -> List[Dict]:
        res_ctx = self.res_ctx
        now = datetime.datetime.now()
        configs = bcs_info_injector.inject_configs(
            access_token=res_ctx.access_token,
            project_id=res_ctx.project_id,
            cluster_id=res_ctx.cluster_id,
            namespace_id=res_ctx.namespace_id,
            namespace=res_ctx.namespace,
            creator=res_ctx.username,
            updator=res_ctx.username,
            created_at=now,
            updated_at=now,
            version=res_ctx.show_version.name,
            source_type='template',
        )
        return configs

    def _get_bcs_variables(self) -> Dict[str, str]:
        res_ctx = self.res_ctx
        namespace_id = res_ctx.namespace_id
        sys_variables = bcs_variable.collect_system_variable(
            access_token=res_ctx.access_token,
            project_id=res_ctx.project_id,
            namespace_id=namespace_id,
        )
        bcs_variables = bcs_variable.get_bcs_variables(res_ctx.project_id, res_ctx.cluster_id, namespace_id)
        sys_variables.update(bcs_variables)
        return sys_variables

    def _render_with_variables(self, raw_content: str, bcs_variables: Dict[str, str]) -> str:
        t = jinja2.Template(raw_content)
        return t.render(bcs_variables)

    def _inject_bcs_info(self, manifest: str, inject_configs: List[Dict]) -> List[Dict]:
        """注入系统配置"""
        # parse_manifest按照yaml分隔符---分割成列表
        manifest_list = bcs_info_injector.parse_manifest(manifest)
        context = {
            'creator': self.res_ctx.username,
            'updator': self.res_ctx.username,
            'version': self.res_ctx.show_version.name,
        }
        manager = bcs_info_injector.InjectManager(configs=inject_configs, resources=manifest_list, context=context)
        return manager.do_inject()

    def _get_raw_manifests(self) -> List[str]:
        """基于资源kind和表id，生成原始的manifest列表"""
        raw_manifests = []
        # instance_entity like {"Deployment": [1, 2]}
        for res_kind, res_file_ids in self.res_ctx.instance_entity.items():
            res_file = res2files.get_resource_file(res_kind, res_file_ids, 'content')
            raw_manifests.extend([f['content'] for f in res_file['files']])
        return raw_manifests

    def _set_namespace(self, manifest: Dict):
        """给NamespaceScoped资源注入指定的命名空间"""
        try:
            # 集群域的资源不指定namespace
            if manifest.get('kind') not in ClusterScopedResources:
                manifest["metadata"]["namespace"] = self.res_ctx.namespace
        except Exception as e:
            raise ParseError(f"set namespace failed: {e}")

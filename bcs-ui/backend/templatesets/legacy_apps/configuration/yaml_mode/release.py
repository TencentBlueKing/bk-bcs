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
from dataclasses import dataclass

import jinja2
import yaml
from rest_framework.exceptions import ParseError

from backend.helm.app import bcs_info_injector
from backend.helm.helm import bcs_variable

from ..constants import FileResourceName
from ..models import ShowVersion


@dataclass
class ReleaseData:
    project_id: str
    namespace_info: dict
    show_version: ShowVersion
    template_files: list
    template_variables: dict


class ReleaseDataProcessor:
    def __init__(self, user, raw_release_data):
        self.access_token = user.token.access_token
        self.username = user.username

        self.project_id = raw_release_data.project_id
        self.namespace_info = raw_release_data.namespace_info
        self.show_version = raw_release_data.show_version
        self.template_files = raw_release_data.template_files
        self.template_variables = raw_release_data.template_variables

    def _get_bcs_variables(self):
        sys_variables = bcs_variable.collect_system_variable(
            access_token=self.access_token, project_id=self.project_id, namespace_id=self.namespace_info["id"]
        )
        bcs_variables = bcs_variable.get_bcs_variables(
            self.project_id, self.namespace_info["cluster_id"], self.namespace_info["id"]
        )
        sys_variables.update(bcs_variables)
        return sys_variables

    def _render_with_variables(self, raw_content, bcs_variables):
        t = jinja2.Template(raw_content)
        return t.render(bcs_variables)

    def _set_namespace(self, resources):
        ignore_ns_res = [
            FileResourceName.ClusterRole.value,
            FileResourceName.ClusterRoleBinding.value,
            FileResourceName.StorageClass.value,
            FileResourceName.PersistentVolume.value,
        ]

        try:
            for res_manifest in resources:
                if res_manifest["kind"] in ignore_ns_res:
                    continue

                metadata = res_manifest["metadata"]
                metadata["namespace"] = self.namespace_info["name"]
        except Exception:
            raise ParseError("set namespace failed: no valid metadata in manifest")

    def _inject_bcs_info(self, yaml_content, inject_configs):
        resources = bcs_info_injector.parse_manifest(yaml_content)
        context = {"creator": self.username, "updator": self.username, "version": self.show_version.name}
        manager = bcs_info_injector.InjectManager(configs=inject_configs, resources=resources, context=context)
        resources = manager.do_inject()
        self._set_namespace(resources)
        return bcs_info_injector.join_manifest(resources)

    def _get_inject_configs(self):
        now = datetime.datetime.now()
        configs = bcs_info_injector.inject_configs(
            access_token=self.access_token,
            project_id=self.project_id,
            cluster_id=self.namespace_info["cluster_id"],
            namespace_id=self.namespace_info["id"],
            namespace=self.namespace_info["name"],
            creator=self.username,
            updator=self.username,
            created_at=now,
            updated_at=now,
            version=self.show_version.name,
            source_type="template",
        )
        return configs

    def _inject(self, raw_content, inject_configs, bcs_variables):
        try:
            content = self._render_with_variables(raw_content, bcs_variables)
            return self._inject_bcs_info(content, inject_configs)
        except Exception as e:
            raise ParseError(f"inject failed: {e}")

    def release_data(self, is_preview=False):
        inject_configs = self._get_inject_configs()
        bcs_variables = self._get_bcs_variables()

        if self.template_variables:
            bcs_variables.update(self.template_variables)

        for res_files in self.template_files:
            for f in res_files["files"]:
                content = self._inject(f["content"], inject_configs, bcs_variables)
                # NOTE: 多转换一次，目的去掉yaml key 上的双引号
                if is_preview:
                    content = yaml.dump(yaml.load(content))
                f["content"] = content
        return ReleaseData(
            self.project_id, self.namespace_info, self.show_version, self.template_files, self.template_variables
        )

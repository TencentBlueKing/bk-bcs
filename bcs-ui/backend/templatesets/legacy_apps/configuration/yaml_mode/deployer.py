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
from backend.bcs_web.audit_log.audit.decorators import log_audit
from backend.bcs_web.audit_log.constants import ActivityType
from backend.utils.client import KubectlClient

from ..auditor import TemplatesetAuditor


class DeployController:
    def __init__(self, user, release_data):
        self.username = user.username
        self.release_data = release_data
        self.namespace = release_data.namespace_info['name']

        self.kubectl = KubectlClient(
            user.token.access_token, release_data.project_id, release_data.namespace_info['cluster_id']
        )

    def _update_audit_ctx(self, activity_type: str):
        show_version = self.release_data.show_version
        template = show_version.related_template

        extra = {}
        for res_file in self.release_data.template_files:
            files_ids = [str(f['id']) for f in res_file['files']]
            extra[res_file['resource_name']] = ','.join(files_ids)

        self.audit_ctx.update_fields(
            project_id=self.release_data.project_id,
            user=self.username,
            activity_type=activity_type,
            resource=template.name,
            resource_id=template.id,
            description=f'deploy template [{template.name}] in ns [{self.namespace}]',
            extra=extra,
        )

    def _to_manifests(self):
        template_files = self.release_data.template_files
        manifest_list = []
        for res_file in template_files:
            manifest_list += [f['content'] for f in res_file['files']]
        return '---\n'.join(manifest_list)

    @log_audit(TemplatesetAuditor)
    def _run_with_kubectl(self, operation):
        manifests = self._to_manifests()
        if operation == 'apply':
            self._update_audit_ctx(activity_type=ActivityType.Instantiate)
            self.kubectl.apply(self.namespace, manifests)
        elif operation == 'delete':
            self._update_audit_ctx(activity_type=ActivityType.Delete)
            self.kubectl.delete(self.namespace, manifests)

    def apply(self):
        self._run_with_kubectl('apply')

    def delete(self):
        self._run_with_kubectl('delete')

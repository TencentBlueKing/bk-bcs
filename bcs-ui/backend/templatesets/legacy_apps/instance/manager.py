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
from typing import Dict

from django.db import models


class VersionInstanceManager(models.Manager):
    """版本实例记录的管理"""

    def update_versions(self, instance_id: int, show_version_name: str, version_id: int, show_version_id: int):
        """更新实例关联的版本信息"""
        self.filter(id=instance_id).update(
            version_id=version_id,
            show_version_id=show_version_id,
            show_version_name=show_version_name,
        )


class InstanceConfigManager(models.Manager):
    """资源实例配置管理"""

    def update_vars_and_configs(self, id: int, variables: Dict, manifest: Dict):
        instance = self.get(id=id)
        instance.variables = json.dumps(variables)
        # 存储格式兼容历史逻辑
        instance.last_config = json.dumps({"old_conf": instance.config})
        instance.config = json.dumps(manifest)
        instance.save(update_fields=["variables", "last_config", "config"])

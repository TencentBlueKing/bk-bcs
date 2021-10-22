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
from typing import Dict

from backend.utils.basic import getitems


def parse_cobj_api_version(crd: Dict) -> str:
    """
    根据 CRD 配置解析 cobj api_version
    ref: https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definition-versioning/#specify-multiple-versions  # noqa
    """
    group = getitems(crd, 'spec.group')
    versions = getitems(crd, 'spec.versions')

    if versions:
        for v in versions:
            if v['served']:
                return f"{group}/{v['name']}"
        return f"{group}/{versions[0]['name']}"

    version = getitems(crd, 'spec.version')
    if version:
        return f"{group}/{version}"

    return f"{group}/v1alpha1"

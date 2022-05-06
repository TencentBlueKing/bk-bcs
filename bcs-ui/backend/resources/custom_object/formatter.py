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

from backend.resources.custom_object.utils import parse_cobj_api_version
from backend.resources.utils.format import ResourceDefaultFormatter
from backend.utils.basic import getitems


class CRDFormatter(ResourceDefaultFormatter):
    def format_dict(self, resource_dict: Dict) -> Dict:
        res = self.format_common_dict(resource_dict)
        res.update(
            {
                'name': getitems(resource_dict, 'metadata.name'),
                'scope': getitems(resource_dict, 'spec.scope'),
                'kind': getitems(resource_dict, 'spec.names.kind'),
                'api_version': parse_cobj_api_version(resource_dict),
            }
        )
        return res


class CustomObjectFormatter(ResourceDefaultFormatter):
    def format_dict(self, resource_dict: Dict) -> Dict:
        return resource_dict


class CustomObjectCommonFormatter(ResourceDefaultFormatter):
    """通用的自定义对象格式化器"""

    def format_dict(self, resource_dict: Dict) -> Dict:
        return self.format_common_dict(resource_dict)

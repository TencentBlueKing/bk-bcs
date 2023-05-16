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
from typing import Dict, List, Union

from backend.resources.constants import PersistentVolumeAccessMode
from backend.resources.storages.common.formatter import StorageFormatter
from backend.utils.basic import getitems


class PersistentVolumeFormatter(StorageFormatter):
    """PersistentVolume 格式化"""

    def parse_access_modes(self, resource_dict: Dict) -> List:
        """access modes 转 缩写用于展示"""
        return [PersistentVolumeAccessMode(m).shortname for m in getitems(resource_dict, 'spec.accessModes', [])]

    def parse_claim(self, resource_dict: Dict) -> Union[str, None]:
        claim_info = getitems(resource_dict, 'spec.claimRef')
        return f"{claim_info['namespace']}/{claim_info['name']}" if claim_info else None

    def format_dict(self, resource_dict: Dict) -> Dict:
        res = self.format_common_dict(resource_dict)

        res.update(
            {
                'accessModes': self.parse_access_modes(resource_dict),
                'claim': self.parse_claim(resource_dict),
            }
        )
        return res

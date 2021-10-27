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
import base64
import gzip
import json
from typing import Dict

from backend.resources.configs.secret.formatter import SecretsFormatter
from backend.utils.basic import getitems


class ReleaseSecretFormatter(SecretsFormatter):
    def format_dict(self, resource_dict: Dict) -> Dict:
        release_data = getitems(resource_dict, "data.release", "")
        if not release_data:
            return {}
        # 解析release data
        release_data = self.parse_release_data(release_data)

        return release_data

    def parse_release_data(self, release_data: bytes) -> Dict:
        """解析release data
        解析release data数据，需要两次 base64，然后 gzip 解压，最后json处理
        ref: https://github.com/helm/helm/blob/main/pkg/storage/driver/secrets.go#L95
        https://github.com/helm/helm/blob/main/pkg/storage/driver/util.go#L56
        """
        # 两次base64解码
        release_data = base64.b64decode(base64.b64decode(release_data))
        # gzip解压
        release_data = gzip.decompress(release_data).decode("utf8")
        # json处理
        return json.loads(release_data)

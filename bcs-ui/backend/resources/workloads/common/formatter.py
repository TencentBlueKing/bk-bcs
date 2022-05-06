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
from typing import Dict, List

from backend.resources.utils.format import ResourceDefaultFormatter
from backend.utils.basic import getitems


class WorkloadFormatter(ResourceDefaultFormatter):
    """工作负载类 资源通用格式化器"""

    def parse_container_images(self, resource_dict: Dict) -> List:
        """
        由 资源配置信息 获取使用的镜像

        :param resource_dict: k8s API 执行结果
        :return: 当前资源容器使用的镜像列表
        """
        containers = getitems(resource_dict, 'spec.template.spec.containers', [])
        return list({c['image'] for c in containers if 'image' in c})

    def format_common_dict(self, resource_dict: Dict) -> Dict:
        ret = super().format_common_dict(resource_dict)
        ret.update({'images': self.parse_container_images(resource_dict)})
        return ret
